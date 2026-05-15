import psycopg2
import time
import csv
import numpy as np
import matplotlib.pyplot as plt
from contextlib import contextmanager

DB_CONFIG = {
    "host": "localhost",
    "database": "zvideo",
    "user": "postgres",
    "password": "1488",
    "port": 5432
}

DATA_SIZES = [10_000, 50_000, 100_000, 250_000, 500_000, 1_000_000]
WARMUP_ITERATIONS = 3
MEASURE_ITERATIONS = 10
TEST_CHANNEL_ID = 1

# Стратегии индексов (выполняются по одной, остальные индексы отсутствуют)
STRATEGIES = {
    "No Index": [],
    "B-Tree on sub(channel_id)": [
        "CREATE INDEX idx_sub_channel_btree ON subscriptions USING btree (channel_id)"
    ],
    "Hash on sub(channel_id)": [
        "CREATE INDEX idx_sub_channel_hash ON subscriptions USING hash (channel_id)"
    ],
    "B-Tree on sub(channel_id, user_id)": [
        "CREATE INDEX idx_sub_channel_user_btree ON subscriptions (channel_id, user_id)"
    ],
    "BRIN on sub(channel_id)": [
        "CREATE INDEX idx_sub_channel_brin ON subscriptions USING brin (channel_id)"
    ],
    "B-Tree on users(notif_enabled, id)": [
        "CREATE INDEX idx_users_notif_id ON users (notifications_enabled, id)"
    ],
    "Covering sub(channel_id) INCLUDE (user_id, new_videos_count)": [
        "CREATE INDEX idx_sub_channel_covering ON subscriptions (channel_id) INCLUDE (user_id, new_videos_count)"
    ]
}

INDEX_NAMES = [
    "idx_sub_channel_btree",
    "idx_sub_channel_hash",
    "idx_sub_channel_user_btree",
    "idx_sub_channel_brin",
    "idx_users_notif_id",
    "idx_sub_channel_covering"
]

def run_benchmark():
    final_data = []

    for size in DATA_SIZES:
        print(f"\n========== DATA SIZE = {size} ==========")
        for strategy_name, index_ddls in STRATEGIES.items():
            print(f"  ▶ Testing: {strategy_name} ...", end=" ", flush=True)

            # Полная перегенерация данных для честных условий
            conn = psycopg2.connect(**DB_CONFIG)
            conn.autocommit = False
            cur = conn.cursor()

            # Очистка и наполнение
            drop_strategy_indexes(cur)
            prepare_database(cur, size)

            # Создание индексов для стратегии
            for ddl in index_ddls:
                cur.execute(ddl)
            cur.execute("ANALYZE users;")
            cur.execute("ANALYZE subscriptions;")
            conn.commit()

            # Счётчик всегда начинается с 0 (DEFAULT)
            times = []
            total_iterations = WARMUP_ITERATIONS + MEASURE_ITERATIONS
            for i in range(total_iterations):
                conn.autocommit = True
                with conn.cursor() as c:
                    c.execute("BEGIN;")
                    start = time.perf_counter()
                    c.execute(f"SELECT notify_subscribers_about_new_video({TEST_CHANNEL_ID});")
                    end = time.perf_counter()
                    c.execute("ROLLBACK;")
                if i >= WARMUP_ITERATIONS:
                    times.append((end - start) * 1000)

            conn.close()

            avg_time = np.mean(times)
            std_dev = np.std(times)
            final_data.append({
                "size": size,
                "strategy": strategy_name,
                "avg_time_ms": avg_time,
                "std_dev_ms": std_dev
            })
            print(f"OK  (avg: {avg_time:.3f} ms, std: {std_dev:.3f} ms)")

    save_results(final_data)

def prepare_database(cur, size):
    """Полная перезагрузка таблиц с заданным числом пользователей и подписок."""
    cur.execute("TRUNCATE users, channels, subscriptions, roles CASCADE;")
    cur.execute("ALTER SEQUENCE users_id_seq RESTART WITH 1;")
    cur.execute("ALTER SEQUENCE channels_id_seq RESTART WITH 1;")

    # Роль для пользователей
    cur.execute("INSERT INTO roles (name, is_default) VALUES ('user', true) RETURNING id;")
    role_id = cur.fetchone()[0]

    # Пользователи: чётные идентификаторы -> notifications_enabled = TRUE
    cur.execute(
        """
        INSERT INTO users (username, email, password_hash, role_id, notifications_enabled)
        SELECT 'user_' || i,
               'email_' || i || '@test.com',
               'hash',
               %s,
               (i %% 2 = 0)
        FROM generate_series(1, %s) AS i
        """,
        (role_id, size)
    )

    # Каналов больше, чтобы фильтр по channel_id был селективным
    channel_count = max(10, size // 100)
    cur.execute(
        """
        INSERT INTO channels (user_id, name)
        SELECT id, 'Channel_' || id
        FROM users
        ORDER BY id
        LIMIT %s
        """,
        (channel_count,)
    )

    # Фиксируем seed для воспроизводимости
    cur.execute("SELECT setseed(0.42);")

    # По одной подписке на пользователя на случайный канал
    cur.execute(
        """
        INSERT INTO subscriptions (user_id, channel_id)
        SELECT id,
               (floor(random() * %s) + 1)::int
        FROM users
        """,
        (channel_count,)
    )

    # Не делаем ANALYZE здесь – он будет после создания индексов в тесте
    print(f"(data generated: {size} users, {channel_count} channels, {size} subs) ", end="", flush=True)

def save_results(data):
    # CSV
    with open("benchmark_results.csv", "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=["size", "strategy", "avg_time_ms", "std_dev_ms"])
        writer.writeheader()
        writer.writerows(data)

    # График
    plt.figure(figsize=(14, 8))
    strategies = list(STRATEGIES.keys())
    for strat in strategies:
        subset = [d for d in data if d["strategy"] == strat]
        if not subset:
            continue
        x = [d["size"] for d in subset]
        y = [d["avg_time_ms"] for d in subset]
        e = [d["std_dev_ms"] for d in subset]
        plt.errorbar(x, y, yerr=e, label=strat, marker="o", capsize=5, linewidth=2)

    plt.xscale("log")
    plt.yscale("log")
    plt.xlabel("Number of subscriptions (data size)", fontsize=12)
    plt.ylabel("Average execution time (ms)", fontsize=12)
    plt.title("Performance of notify_subscribers_about_new_video()\nwith different indexing strategies", fontsize=14)
    plt.legend(bbox_to_anchor=(1.05, 1), loc="upper left")
    plt.grid(True, which="both", linestyle="--", alpha=0.7)
    plt.tight_layout()
    plt.savefig("benchmark_plot.png", dpi=150)
    print("\nResults saved to benchmark_results.csv and benchmark_plot.png")

def drop_strategy_indexes(cur):
    # Удаляем только индексы из эксперимента, чтобы изоляция была корректной
    for index_name in INDEX_NAMES:
        cur.execute(f"DROP INDEX IF EXISTS {index_name};")

if __name__ == "__main__":
    run_benchmark()