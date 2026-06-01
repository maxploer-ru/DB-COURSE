import csv
import time
import os
import sys
import numpy as np
import matplotlib.pyplot as plt
import psycopg2
from contextlib import contextmanager

DB_CONFIG = {
    "host": "localhost",
    "database": "zvideo",
    "user": "postgres",
    "password": "1488",
    "port": 5432,
    "connect_timeout": 30,
    "options": "-c statement_timeout=300000 -c idle_in_transaction_session_timeout=120000"
}

# Размеры данных для тестирования
DATA_SIZES = [
    10_000, 50_000,
    100_000, 250_000,
    500_000, 750_000, 1_000_000
]

# Параметры бенчмарка
WARMUP_ITERATIONS = 5
MEASURE_ITERATIONS = 20
TEST_CHANNEL_ID = 1

# Доли подписчиков тестового канала
TARGET_CHANNEL_FRACTIONS = [0.001, 0.01, 0.05, 0.10]

# Если стандартное отклонение слишком велико, повторяем замер
STDDEV_MAX_RATIO = 0.1  # 20% от среднего
MAX_STDDEV_RETRIES = 100000

# Стратегии индексов
STRATEGIES = {
    "No Index": [],
    "B-Tree on sub(channel_id)": [
        "CREATE INDEX IF NOT EXISTS idx_sub_channel_btree ON subscriptions USING btree (channel_id)"
    ],
    "B-Tree on sub(channel_id, user_id)": [
        "CREATE INDEX IF NOT EXISTS idx_sub_channel_user_btree ON subscriptions USING btree (channel_id, user_id)"
    ],
    "B-Tree on users(notif_enabled, id)": [
        "CREATE INDEX IF NOT EXISTS idx_users_notif_id_btree ON users USING btree (notifications_enabled, id)"
    ],
    "B-Tree partial users(id) WHERE notif_enabled": [
        "CREATE INDEX IF NOT EXISTS idx_users_active_btree ON users USING btree (id) WHERE notifications_enabled = TRUE"
    ],
    "Combined: sub(channel_id) + users(notif_enabled, id)": [
        "CREATE INDEX IF NOT EXISTS idx_sub_channel_btree ON subscriptions USING btree (channel_id)",
        "CREATE INDEX IF NOT EXISTS idx_users_notif_id_btree ON users USING btree (notifications_enabled, id)"
    ]
}

ALL_INDEX_NAMES = [
    "idx_sub_channel_btree",
    "idx_sub_channel_user_btree",
    "idx_users_notif_id_btree",
    "idx_users_active_btree"
]

plt.rcParams["hatch.linewidth"] = 1.2
plt.rcParams["legend.handlelength"] = 3.5
plt.rcParams["legend.handleheight"] = 1.2

PLOT_COLORS = [
    "#4C78A8",
    "#F58518",
    "#54A24B",
    "#E45756",
    "#72B7B2",
    "#B279A2",
    "#FF9DA6",
    "#9D755D",
    "#BAB0AC",
]

PLOT_HATCHES = ["/", "\\\\", "|", "-", "x", ".", "o"]

OUTPUT_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "report", "images"))


@contextmanager
def get_connection():
    """Контекстный менеджер для подключения к БД"""
    conn = psycopg2.connect(**DB_CONFIG)
    try:
        yield conn
    finally:
        conn.close()


def kill_idle_transactions(cur):
    """Убиваем подвисшие idle транзакции, которые блокируют VACUUM"""
    cur.execute("""
                SELECT pg_terminate_backend(pid)
                FROM pg_stat_activity
                WHERE pid <> pg_backend_pid()
                  AND state = 'idle in transaction'
                  AND query NOT LIKE '%pg_stat_activity%';
                """)


def drop_all_experiment_indexes(cur):
    """Удаление всех индексов эксперимента"""
    for index_name in ALL_INDEX_NAMES:
        cur.execute(f"DROP INDEX IF EXISTS {index_name};")


def create_indexes(cur, index_ddls):
    """Создание индексов для стратегии"""
    for ddl in index_ddls:
        cur.execute(ddl)


def vacuum_all_tables(conn, full=False):
    """Тщательная очистка всех таблиц с VACUUM"""
    vacuum_type = "VACUUM FULL" if full else "VACUUM"

    # Завершаем любые открытые транзакции перед переключением autocommit
    try:
        conn.rollback()
    except Exception:
        pass

    # Временно включаем autocommit, так как VACUUM нельзя выполнять в транзакции
    old_autocommit = conn.autocommit
    conn.autocommit = True

    try:
        with conn.cursor() as cur:
            # Убиваем мешающие транзакции
            kill_idle_transactions(cur)

            for table in ["users", "subscriptions", "channels", "roles"]:
                try:
                    cur.execute(f"{vacuum_type} ANALYZE {table};")
                    print(f"     ✓ VACUUM {table} completed")
                except Exception as e:
                    print(f"     ⚠ Warning: VACUUM {table} failed: {e}")
    finally:
        conn.autocommit = old_autocommit


def prepare_database(conn, size, target_fraction):
    """
    Полная перегенерация данных с контролируемым распределением подписок.
    Всё выполняется в отдельных транзакциях для предотвращения блокировок.
    """

    # === Этап 1: Очистка таблиц (отдельная транзакция) ===
    with conn.cursor() as cur:
        cur.execute("DROP TABLE IF EXISTS subscriptions CASCADE;")
        cur.execute("DROP TABLE IF EXISTS channels CASCADE;")
        cur.execute("DROP TABLE IF EXISTS users CASCADE;")
        cur.execute("DROP TABLE IF EXISTS roles CASCADE;")
    conn.commit()

    # === Этап 2: Создание таблиц заново (чтобы избежать фрагментации) ===
    with conn.cursor() as cur:
        # roles
        cur.execute("""
                    CREATE TABLE roles (
                                           id SERIAL PRIMARY KEY,
                                           name VARCHAR(32) UNIQUE NOT NULL,
                                           is_default BOOLEAN NOT NULL DEFAULT FALSE
                    );
                    """)
        cur.execute("""
                    INSERT INTO roles (name, is_default) VALUES
                                                             ('admin', false), ('moderator', false), ('user', true);
                    """)
        cur.execute("SELECT id FROM roles WHERE name = 'user';")
        role_id = cur.fetchone()[0]
    conn.commit()

    # === Этап 3: Создание и заполнение users (отдельная транзакция) ===
    with conn.cursor() as cur:
        cur.execute("""
                    CREATE TABLE users (
                                           id SERIAL PRIMARY KEY,
                                           role_id INT NOT NULL REFERENCES roles (id) ON DELETE RESTRICT,
                                           username VARCHAR(32) UNIQUE NOT NULL,
                                           email VARCHAR(64) UNIQUE NOT NULL,
                                           password_hash TEXT NOT NULL,
                                           is_active BOOLEAN NOT NULL DEFAULT TRUE,
                                           notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
                                           created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                           updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
                    );
                    """)

        cur.execute("""
                    INSERT INTO users (username, email, password_hash, role_id, notifications_enabled)
                    SELECT
                        'user_' || id,
                        'email_' || id || '@test.com',
                        'hash',
                        %s,
                        (id %% 2 = 0)
                    FROM generate_series(1, %s) AS id;
                    """, (role_id, size))
    conn.commit()

    # === Этап 4: Создание и заполнение channels (отдельная транзакция) ===
    with conn.cursor() as cur:
        cur.execute("""
                    CREATE TABLE channels (
                                              id SERIAL PRIMARY KEY,
                                              user_id INT UNIQUE NOT NULL REFERENCES users (id) ON DELETE CASCADE,
                                              name VARCHAR(32) UNIQUE NOT NULL,
                                              description TEXT,
                                              created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
                    );
                    """)

        channel_count = max(10, size // 10)
        cur.execute("""
                    INSERT INTO channels (user_id, name)
                    SELECT id, 'Channel_' || id
                    FROM users
                    ORDER BY id
                    LIMIT %s;
                    """, (channel_count,))
    conn.commit()

    # === Этап 5: Создание и заполнение subscriptions (отдельная транзакция) ===
    with conn.cursor() as cur:
        cur.execute("""
                    CREATE TABLE subscriptions (
                                                   user_id INT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
                                                   channel_id INT NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
                                                   PRIMARY KEY (user_id, channel_id),
                                                   new_videos_count INT NOT NULL DEFAULT 0,
                                                   subscribed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
                    );
                    """)

        # Фиксируем seed для воспроизводимости
        cur.execute("SELECT setseed(0.42);")

        # Пороги для распределения (сумма = 1.0)
        frac_target = target_fraction
        frac_popular = 0.20                    # 20% → каналы 2-10
        frac_medium = 0.30                     # 30% → каналы 11-100
        # остальное (50% - frac_target) → каналы 101+

        # Условные вероятности
        t1 = frac_target
        t2 = frac_popular / (1.0 - t1)
        t3 = frac_medium / (1.0 - t1 - frac_popular)

        cur.execute(f"""
            INSERT INTO subscriptions (user_id, channel_id)
            SELECT
                u.id,
                CASE
                    WHEN random() < {t1} THEN {TEST_CHANNEL_ID}
                    WHEN random() < {t2} THEN (floor(random() * 9) + 2)::int
                    WHEN random() < {t3} THEN (floor(random() * 90) + 11)::int
                    ELSE (floor(random() * GREATEST({channel_count} - 100, 1)) + 101)::int
                END
            FROM users u;
        """)
    conn.commit()

    # === Этап 6: VACUUM после массовой вставки (вне транзакции!) ===
    print("     Running VACUUM ANALYZE after data generation...")
    vacuum_all_tables(conn, full=False)

    # === Этап 7: Подсчёт affected rows для тестового канала ===
    with conn.cursor() as cur:
        cur.execute("""
                    SELECT COUNT(*)
                    FROM subscriptions s
                             JOIN users u ON s.user_id = u.id
                    WHERE s.channel_id = %s AND u.notifications_enabled = TRUE;
                    """, (TEST_CHANNEL_ID,))
        affected_rows = cur.fetchone()[0]

    return channel_count, affected_rows


def reset_counters(conn, channel_id):
    """Сброс счётчиков новых видео (отдельная транзакция)"""
    with conn.cursor() as cur:
        cur.execute("""
                    UPDATE subscriptions
                    SET new_videos_count = 0
                    WHERE channel_id = %s;
                    """, (channel_id,))
    conn.commit()


def benchmark_function(conn, channel_id):
    """Измерение времени выполнения процедуры с реальным COMMIT"""
    start = time.perf_counter()

    with conn.cursor() as cur:
        cur.execute("CALL notify_subscribers_about_new_video(%s);", (channel_id,))

    conn.commit()

    elapsed = (time.perf_counter() - start) * 1000
    return elapsed


def measure_with_retries(conn, channel_id):
    """Замер с повтором при большом стандартном отклонении"""
    last_stats = None
    for attempt in range(1, MAX_STDDEV_RETRIES + 1):
        # Прогрев
        for _ in range(WARMUP_ITERATIONS):
            benchmark_function(conn, channel_id)
            reset_counters(conn, channel_id)

        times = []
        for i in range(MEASURE_ITERATIONS):
            elapsed = benchmark_function(conn, channel_id)
            times.append(elapsed)
            reset_counters(conn, channel_id)

            if (i + 1) % 5 == 0:
                print(".", end="", flush=True)

        avg_time = np.mean(times)
        std_dev = np.std(times)
        median_time = np.median(times)
        min_time = np.min(times)
        max_time = np.max(times)

        last_stats = {
            "avg_time_ms": avg_time,
            "std_dev_ms": std_dev,
            "median_time_ms": median_time,
            "min_time_ms": min_time,
            "max_time_ms": max_time,
        }

        if avg_time > 0 and (std_dev / avg_time) <= STDDEV_MAX_RATIO:
            return last_stats

        if attempt < MAX_STDDEV_RETRIES:
            print(" повтор", end="", flush=True)

    return last_stats


def run_benchmark():
    """Основная функция бенчмарка"""
    final_data = []

    for target_fraction in TARGET_CHANNEL_FRACTIONS:
        print(f"\n{'='*80}")
        print(f"TARGET_CHANNEL_FRACTION = {target_fraction}")
        print(f"{'='*80}")

        for size in DATA_SIZES:
            print(f"\n{'='*50}")
            print(f"DATA SIZE = {size:,}")
            print(f"{'='*50}")

            # Для каждого размера один раз готовим данные
            print("  Preparing database...")
            with get_connection() as conn:
                channel_count, _ = prepare_database(conn, size, target_fraction)
                print(f"  Channels: {channel_count}, continuing with strategies...")

            for strategy_name, index_ddls in STRATEGIES.items():
                print(f"  ▶ {strategy_name:<50} ...", end=" ", flush=True)

                with get_connection() as conn:
                    with conn.cursor() as cur:
                        # Очистка старых индексов
                        drop_all_experiment_indexes(cur)

                        # Создание индексов для текущей стратегии
                        create_indexes(cur, index_ddls)

                        # Обновление статистики
                        cur.execute("ANALYZE users;")
                        cur.execute("ANALYZE subscriptions;")
                    conn.commit()

                    # Получаем актуальное число affected rows
                    with conn.cursor() as cur:
                        cur.execute("""
                                    SELECT COUNT(*)
                                    FROM subscriptions s
                                             JOIN users u ON s.user_id = u.id
                                    WHERE s.channel_id = %s AND u.notifications_enabled = TRUE;
                                    """, (TEST_CHANNEL_ID,))
                        affected_rows = cur.fetchone()[0]

                    # Измерение с VACUUM между стратегиями для чистоты
                    if strategy_name != list(STRATEGIES.keys())[0]:
                        print("v", end="", flush=True)
                        vacuum_all_tables(conn, full=False)

                    stats = measure_with_retries(conn, TEST_CHANNEL_ID)

                final_data.append({
                    "size": size,
                    "strategy": strategy_name,
                    "avg_time_ms": stats["avg_time_ms"],
                    "std_dev_ms": stats["std_dev_ms"],
                    "median_time_ms": stats["median_time_ms"],
                    "min_time_ms": stats["min_time_ms"],
                    "max_time_ms": stats["max_time_ms"],
                    "affected_rows": affected_rows,
                    "channel_count": channel_count,
                    "target_fraction": target_fraction,
                })

                print(" OK")
                print(f"     Affected rows: {affected_rows:,}")
                print(f"     Avg: {stats['avg_time_ms']:.3f} ms, Median: {stats['median_time_ms']:.3f} ms")
                print(f"     Std: {stats['std_dev_ms']:.3f} ms, Min: {stats['min_time_ms']:.3f} ms, Max: {stats['max_time_ms']:.3f} ms")

    save_results(final_data)


def save_results(data):
    """Сохранение результатов в CSV и построение графиков"""
    csv_file = "../../benchmark_results.csv"
    with open(csv_file, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=[
            "size", "strategy", "avg_time_ms", "std_dev_ms",
            "median_time_ms", "min_time_ms", "max_time_ms",
            "affected_rows", "channel_count", "target_fraction"
        ])
        writer.writeheader()
        writer.writerows(data)

    print(f"\n✓ Results saved to {csv_file}")

    for target_fraction in TARGET_CHANNEL_FRACTIONS:
        subset = [d for d in data if d["target_fraction"] == target_fraction]
        if not subset:
            continue
        plot_results(subset, target_fraction)
        plot_relative_performance(subset, target_fraction)


def plot_results(data, target_fraction):
    """Основная диаграмма (абсолютные значения)"""
    plt.figure(figsize=(16, 8))
    strategies = list(STRATEGIES.keys())
    sizes = sorted(set(d["size"] for d in data))
    x = np.arange(len(sizes))
    width = 0.8 / max(len(strategies), 1)

    size_to_idx = {size: idx for idx, size in enumerate(sizes)}

    for idx, strat in enumerate(strategies):
        subset = [d for d in data if d["strategy"] == strat]
        if not subset:
            continue
        values = [0.0] * len(sizes)
        for row in subset:
            values[size_to_idx[row["size"]]] = row["avg_time_ms"]
        offset = (idx - (len(strategies) - 1) / 2) * width
        plt.bar(
            x + offset,
            values,
            width=width,
            label=strat,
            color=PLOT_COLORS[idx % len(PLOT_COLORS)],
            hatch=PLOT_HATCHES[idx % len(PLOT_HATCHES)],
            edgecolor="black",
            linewidth=0.6,
        )

    plt.xticks(x, [f"{s:,}" for s in sizes], rotation=0)
    plt.xlabel("Количество пользователей/подписок", fontsize=13)
    plt.ylabel("Среднее время выполнения (мс)", fontsize=13)
    plt.title(
        f"Производительность notify_subscribers_about_new_video()\n"
        f"(абсолютные значения, доля {target_fraction * 100:.1f}%)",
        fontsize=14, fontweight="bold"
    )
    plt.legend(bbox_to_anchor=(1.05, 1), loc="upper left", fontsize=10)
    plt.grid(True, axis="y", linestyle="--", alpha=0.6)
    plt.tight_layout()
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    suffix = fraction_to_suffix(target_fraction)
    png_path = os.path.join(OUTPUT_DIR, f"benchmark_plot_{suffix}.png")
    pdf_path = os.path.join(OUTPUT_DIR, f"benchmark_plot_{suffix}.pdf")
    plt.savefig(png_path, dpi=150, bbox_inches="tight")
    plt.savefig(pdf_path, bbox_inches="tight")
    print(f"✓ Plot saved to {png_path}")


def plot_relative_performance(data, target_fraction):
    """Относительная диаграмма (нормализация 0..1)"""
    plt.figure(figsize=(16, 8))
    strategies = list(STRATEGIES.keys())
    sizes = sorted(set(d["size"] for d in data))
    x = np.arange(len(sizes))
    width = 0.8 / max(len(strategies), 1)

    by_size = {size: [] for size in sizes}
    for row in data:
        by_size[row["size"]].append(row)

    size_to_max = {}
    for size, rows in by_size.items():
        if rows:
            size_to_max[size] = max(r["avg_time_ms"] for r in rows)
        else:
            size_to_max[size] = 1.0

    size_to_idx = {size: idx for idx, size in enumerate(sizes)}

    for idx, strat in enumerate(strategies):
        subset = [d for d in data if d["strategy"] == strat]
        if not subset:
            continue
        values = [0.0] * len(sizes)
        for row in subset:
            denom = size_to_max[row["size"]] or 1.0
            values[size_to_idx[row["size"]]] = row["avg_time_ms"] / denom
        offset = (idx - (len(strategies) - 1) / 2) * width
        plt.bar(
            x + offset,
            values,
            width=width,
            label=strat,
            color=PLOT_COLORS[idx % len(PLOT_COLORS)],
            hatch=PLOT_HATCHES[idx % len(PLOT_HATCHES)],
            edgecolor="black",
            linewidth=0.6,
        )

    plt.ylim(0, 1.05)
    plt.xticks(x, [f"{s:,}" for s in sizes], rotation=0)
    plt.xlabel("Количество пользователей/подписок", fontsize=13)
    plt.ylabel("Относительное время (доля от максимума)", fontsize=13)
    plt.title(
        f"Относительная производительность (0..1)\n"
        f"(доля {target_fraction * 100:.1f}%)",
        fontsize=14, fontweight="bold"
    )
    plt.legend(bbox_to_anchor=(1.05, 1), loc="upper left", fontsize=10)
    plt.grid(True, axis="y", linestyle="--", alpha=0.6)
    plt.tight_layout()
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    suffix = fraction_to_suffix(target_fraction)
    png_path = os.path.join(OUTPUT_DIR, f"benchmark_relative_{suffix}.png")
    pdf_path = os.path.join(OUTPUT_DIR, f"benchmark_relative_{suffix}.pdf")
    plt.savefig(png_path, dpi=150, bbox_inches="tight")
    plt.savefig(pdf_path, bbox_inches="tight")
    print(f"✓ Relative performance plot saved to {png_path}")



def fraction_to_suffix(value):
    """Безопасный суффикс для имени файла (0.001 -> 0_001)"""
    return f"{str(value).replace('.', '_')}"


def load_results(csv_path):
    with open(csv_path, "r", newline="") as f:
        reader = csv.DictReader(f)
        data = list(reader)
    for row in data:
        row["size"] = int(row["size"])
        row["avg_time_ms"] = float(row["avg_time_ms"])
        row["std_dev_ms"] = float(row["std_dev_ms"])
        row["median_time_ms"] = float(row["median_time_ms"])
        row["min_time_ms"] = float(row["min_time_ms"])
        row["max_time_ms"] = float(row["max_time_ms"])
        row["affected_rows"] = int(row["affected_rows"])
        row["target_fraction"] = float(row["target_fraction"])
    return data


def print_summary(data):
    """Вывод итоговой таблицы"""
    for target_fraction in TARGET_CHANNEL_FRACTIONS:
        subset_fraction = [d for d in data if d.get("target_fraction") == target_fraction]
        if not subset_fraction:
            continue

        print(f"\n{'='*100}")
        print(f"PERFORMANCE SUMMARY (TARGET_CHANNEL_FRACTION={target_fraction})")
        print(f"{'='*100}")

        sizes = sorted(set(d["size"] for d in subset_fraction))

        for size in sizes:
            subset = [d for d in subset_fraction if d["size"] == size]
            subset.sort(key=lambda x: x["avg_time_ms"])

            best = subset[0]
            worst = subset[-1]
            improvement = ((worst["avg_time_ms"] - best["avg_time_ms"]) / worst["avg_time_ms"]) * 100

            print(f"\nSize: {size:,} | Affected rows: {best['affected_rows']:,}")
            print(f"  Best:  {best['strategy']:<45} {best['avg_time_ms']:.3f} ms")
            print(f"  Worst: {worst['strategy']:<45} {worst['avg_time_ms']:.3f} ms")
            print(f"  Improvement: {improvement:.1f}%")

        print(f"\n{'='*100}")
        print("GLOBAL RECOMMENDATION")
        print(f"{'='*100}")

        strategy_avg = {}
        for strat in STRATEGIES.keys():
            subset = [d for d in subset_fraction if d["strategy"] == strat]
            if subset:
                strategy_avg[strat] = np.mean([d["avg_time_ms"] for d in subset])

        best_strategy = min(strategy_avg, key=strategy_avg.get)
        print(f"\nBest overall strategy: {best_strategy}")
        print(f"Average execution time: {strategy_avg[best_strategy]:.3f} ms")


if __name__ == "__main__":
    if "--plot-only" in sys.argv:
        csv_path = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "benchmark_results.csv"))
        data = load_results(csv_path)
        for target_fraction in TARGET_CHANNEL_FRACTIONS:
            subset = [d for d in data if d.get("target_fraction") == target_fraction]
            if not subset:
                continue
            plot_results(subset, target_fraction)
            plot_relative_performance(subset, target_fraction)
        sys.exit(0)

    print("=" * 80)
    print("POSTGRESQL INDEX BENCHMARK: notify_subscribers_about_new_video()")
    print("=" * 80)
    print(f"Test channel ID: {TEST_CHANNEL_ID}")
    print(f"Target channel fractions: {TARGET_CHANNEL_FRACTIONS}")
    print(f"Data sizes: {len(DATA_SIZES)} configurations")
    print(f"Strategies: {len(STRATEGIES)}")
    print(f"Warmup iterations: {WARMUP_ITERATIONS}")
    print(f"Measurement iterations: {MEASURE_ITERATIONS}")
    print(f"Stddev max ratio: {STDDEV_MAX_RATIO}")
    print(f"Max stddev retries: {MAX_STDDEV_RETRIES}")
    print("=" * 80)

    start_time = time.time()

    run_benchmark()

    data = load_results(os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "benchmark_results.csv")))

    print_summary(data)

    elapsed = time.time() - start_time
    print(f"\n{'='*80}")
    print(f"Total benchmark time: {elapsed/60:.1f} minutes")
    print(f"{'='*80}")