package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
)

func listCommentsFlow(videoID int, isOwner bool) error {
	page := 0
	limit := 5
	for {
		var resp CommentListResponse
		path := fmt.Sprintf("/videos/%d/comments?limit=%d&offset=%d", videoID, limit, page*limit)
		err := doRequest("GET", path, nil, &resp)
		if err != nil {
			return err
		}
		if len(resp.Comments) == 0 && page == 0 {
			fmt.Println("Комментариев пока нет.")
			return nil
		}
		if len(resp.Comments) == 0 {
			fmt.Println("Больше нет комментариев.")
			return nil
		}

		items := make([]string, len(resp.Comments)+1)
		for i, c := range resp.Comments {
			timeStr := c.CreatedAt.Format("2006-01-02 15:04")
			items[i] = fmt.Sprintf("[%s] Пользователь %d (нравится %d; не нравится %d): %s",
				timeStr, c.UserID, c.Likes, c.Dislikes, truncate(c.Content, 40))
		}
		items[len(resp.Comments)] = "Следующая страница →"

		label := fmt.Sprintf("Комментарии к видео %d (страница %d)", videoID, page+1)
		selectPrompt := promptui.Select{
			Label: label,
			Items: items,
			Size:  10,
		}
		idx, _, err := selectPrompt.Run()
		if err != nil {
			return err
		}
		if idx == len(resp.Comments) {
			page++
			continue
		}

		selected := resp.Comments[idx]
		fmt.Printf("\n--- Комментарий пользователя %d (нравится %d; не нравится %d) ---\n%s\nОпубликован: %s\n\n",
			selected.UserID, selected.Likes, selected.Dislikes, selected.Content,
			selected.CreatedAt.Format(time.RFC3339))

		actions := []string{"Назад к списку"}
		actions = append(actions, "Редактировать (если ваш)", "Удалить (если разрешено)", "Оценить")

		actionPrompt := promptui.Select{
			Label: "Действие",
			Items: actions,
		}
		actIdx, _, err := actionPrompt.Run()
		if err != nil {
			continue
		}
		switch actIdx {
		case 0:
			continue
		case 1:
			if err := editCommentFlow(selected.ID); err != nil {
				fmt.Printf("Ошибка редактирования: %v\n", err)
			}
		case 2:
			if err := deleteCommentFlow(selected.ID); err != nil {
				fmt.Printf("Ошибка удаления: %v\n", err)
			}
		case 3:
			rateCommentFlow(selected.ID)
		}
	}
}

func addCommentFlow(videoID int) error {
	prompt := promptui.Prompt{
		Label: "Введите ваш комментарий",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("комментарий не может быть пустым")
			}
			return nil
		},
	}
	content, err := prompt.Run()
	if err != nil {
		return err
	}

	req := map[string]string{"content": content}
	err = doRequest("POST", fmt.Sprintf("/videos/%d/comments", videoID), req, nil)
	if err != nil {
		return err
	}
	fmt.Println("Комментарий добавлен.")
	return nil
}

func editCommentFlow(commentID int) error {
	prompt := promptui.Prompt{
		Label: "Введите новый текст комментария",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("комментарий не может быть пустым")
			}
			return nil
		},
	}
	newContent, err := prompt.Run()
	if err != nil {
		return err
	}

	req := map[string]string{"content": newContent}
	err = doRequest("PATCH", fmt.Sprintf("/comments/%d", commentID), req, nil)
	if err != nil {
		return err
	}
	fmt.Println("Комментарий обновлён.")
	return nil
}

func deleteCommentFlow(commentID int) error {
	confirm := promptui.Prompt{
		Label: "Введите 'да' для подтверждения удаления",
		Validate: func(input string) error {
			if input != "да" {
				return fmt.Errorf("подтверждение не получено")
			}
			return nil
		},
	}
	_, err := confirm.Run()
	if err != nil {
		return err
	}
	err = doRequest("DELETE", fmt.Sprintf("/comments/%d", commentID), nil, nil)
	if err != nil {
		return err
	}
	fmt.Println("Комментарий удалён.")
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func rateCommentFlow(commentID int) {
	for {
		actions := []string{
			"Нравится",
			"Не нравится",
			"Убрать оценку",
			"Назад",
		}
		prompt := promptui.Select{
			Label: fmt.Sprintf("Оценка комментария %d", commentID),
			Items: actions,
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return
		}
		switch idx {
		case 0:
			err = doRequest("POST", fmt.Sprintf("/comments/%d/like", commentID), nil, nil)
		case 1:
			err = doRequest("POST", fmt.Sprintf("/comments/%d/dislike", commentID), nil, nil)
		case 2:
			err = doRequest("DELETE", fmt.Sprintf("/comments/%d/rating", commentID), nil, nil)
		case 3:
			return
		}
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
		} else if idx != 3 {
			fmt.Println("Готово")
		}
	}
}
