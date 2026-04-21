package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/manifoldco/promptui"
)

func listAllVideosFlow() error {
	page := 0
	limit := 5
	for {
		var videos []VideoResponse
		err := doRequest("GET", fmt.Sprintf("/videos?limit=%d&offset=%d", limit, page*limit), nil, &videos)
		if err != nil {
			return err
		}
		if len(videos) == 0 && page == 0 {
			fmt.Println("Видео не найдены.")
			return nil
		}
		if len(videos) == 0 {
			fmt.Println("Больше нет видео.")
			return nil
		}
		items := make([]string, len(videos)+1)
		for i, v := range videos {
			items[i] = fmt.Sprintf("%s (просмотров %d; нравится %d; не нравится %d; комментариев %d)",
				v.Title, v.Views, v.Likes, v.Dislikes, v.Comments)
		}
		items[len(videos)] = "Следующая страница →"
		selectPrompt := promptui.Select{
			Label: "Все видео",
			Items: items,
			Size:  10,
		}
		idx, _, err := selectPrompt.Run()
		if err != nil {
			return err
		}
		if idx == len(videos) {
			page++
			continue
		}
		selected := videos[idx]
		watchVideoFlow(selected.ID, selected.ChannelID)
	}
}

func watchVideoFlow(videoID, channelID int) {
	var streamResp StreamURLResponse
	err := doRequest("GET", fmt.Sprintf("/videos/%d/stream-url", videoID), nil, &streamResp)
	if err != nil {
		fmt.Printf("Ошибка получения URL потока: %v\n", err)
		return
	}
	fmt.Printf("URL для просмотра: %s\n", streamResp.URL)
	fmt.Println("(Вы можете открыть эту ссылку в браузере или медиаплеере)")

	for {
		actions := []string{
			"Оценить видео",
			"Просмотреть комментарии",
			"Добавить комментарий",
			"Управление подпиской на канал",
			"Назад к списку видео",
		}
		prompt := promptui.Select{
			Label: "Что вы хотите сделать?",
			Items: actions,
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return
		}
		switch idx {
		case 0:
			rateVideoFlow(videoID)
		case 1:
			if err := listCommentsFlow(videoID, false); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case 2:
			if err := addCommentFlow(videoID); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case 3:
			subscriptionActionFlow(channelID)
		case 4:
			return
		}
	}
}

func videoActionsFlow(videoID, channelID int) {
	for {
		actions := []string{
			"Смотреть",
			"Редактировать название/описание",
			"Оценить",
			"Управление комментариями",
			"Удалить",
			"Назад",
		}
		prompt := promptui.Select{
			Label: fmt.Sprintf("Видео ID %d", videoID),
			Items: actions,
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return
		}
		switch idx {
		case 0:
			watchVideoFlow(videoID, channelID)
		case 1:
			editVideoFlow(videoID)
		case 2:
			rateVideoFlow(videoID)
		case 3:
			if err := listCommentsFlow(videoID, true); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case 4:
			if deleteVideoFlow(videoID) {
				return
			}
		case 5:
			return
		}
	}
}

func rateVideoFlow(videoID int) {
	for {
		actions := []string{
			"Нравится",
			"Не нравится",
			"Убрать оценку",
			"Назад",
		}
		prompt := promptui.Select{
			Label: fmt.Sprintf("Оценка видео %d", videoID),
			Items: actions,
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return
		}
		switch idx {
		case 0:
			err = doRequest("POST", fmt.Sprintf("/videos/%d/like", videoID), nil, nil)
		case 1:
			err = doRequest("POST", fmt.Sprintf("/videos/%d/dislike", videoID), nil, nil)
		case 2:
			err = doRequest("DELETE", fmt.Sprintf("/videos/%d/rating", videoID), nil, nil)
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

func myVideosFlow() error {
	page := 0
	limit := 5
	for {
		var videos []VideoResponse
		err := doRequest("GET", fmt.Sprintf("/videos/me?limit=%d&offset=%d", limit, page*limit), nil, &videos)
		if err != nil {
			return err
		}
		if len(videos) == 0 && page == 0 {
			fmt.Println("У вас нет видео.")
			return nil
		}
		if len(videos) == 0 {
			fmt.Println("Больше нет видео.")
			return nil
		}
		items := make([]string, len(videos)+1)
		for i, v := range videos {
			items[i] = fmt.Sprintf("%s (просмотров %d; нравится %d; не нравится %d; комментариев %d)",
				v.Title, v.Views, v.Likes, v.Dislikes, v.Comments)
		}
		items[len(videos)] = "Следующая страница →"
		prompt := promptui.Select{
			Label: "Ваши видео",
			Items: items,
			Size:  10,
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return err
		}
		if idx == len(videos) {
			page++
			continue
		}
		selected := videos[idx]
		videoActionsFlow(selected.ID, selected.ChannelID)
	}
}

func uploadVideoFlow() error {
	var channel ChannelResponse
	err := doRequest("GET", "/channels/me", nil, &channel)
	if err != nil {
		fmt.Printf("Не удалось получить ваш канал: %v\n", err)
		return nil
	}
	prompt := promptui.Prompt{Label: "Название видео"}
	title, err := prompt.Run()
	if err != nil {
		return err
	}
	prompt = promptui.Prompt{Label: "Описание (необязательно)"}
	desc, _ := prompt.Run()
	prompt = promptui.Prompt{Label: "Путь к локальному файлу"}
	filePath, err := prompt.Run()
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("Ошибка файла: %v\n", err)
		return nil
	}
	if fileInfo.IsDir() {
		fmt.Println("Указан путь к папке, а не к файлу.")
		return nil
	}
	filename := fileInfo.Name()
	reqBody := map[string]interface{}{
		"channel_id": channel.ID,
		"filename":   filename,
	}
	var uploadResp UploadURLResponse
	err = doRequest("POST", "/videos/upload-url", reqBody, &uploadResp)
	if err != nil {
		return err
	}
	fmt.Printf("Загрузка %s (%d байт)...\n", filename, fileInfo.Size())
	err = uploadFileToPresignedURL(uploadResp.URL, filePath)
	if err != nil {
		fmt.Printf("Ошибка загрузки: %v\n", err)
		return nil
	}
	fmt.Println("Файл загружен.")
	createReq := map[string]interface{}{
		"channel_id":  channel.ID,
		"title":       title,
		"description": desc,
		"file_key":    uploadResp.FileKey,
	}
	var video VideoResponse
	err = doRequest("POST", "/videos", createReq, &video)
	if err != nil {
		return err
	}
	fmt.Printf("Видео создано! ID: %d\n", video.ID)
	return nil
}

func uploadFileToPresignedURL(url, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return err
	}
	req.ContentLength = stat.Size()
	req.Header.Set("Content-Type", "application/octet-stream")
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("загрузка вернула статус %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func editVideoFlow(videoID int) {
	prompt := promptui.Prompt{Label: "Новое название (оставьте пустым, чтобы не менять)"}
	newTitle, _ := prompt.Run()
	prompt = promptui.Prompt{Label: "Новое описание (оставьте пустым, чтобы не менять)"}
	newDesc, _ := prompt.Run()
	req := make(map[string]*string)
	if newTitle != "" {
		req["title"] = &newTitle
	}
	if newDesc != "" {
		req["description"] = &newDesc
	}
	if len(req) == 0 {
		return
	}
	err := doRequest("PATCH", fmt.Sprintf("/videos/%d", videoID), req, nil)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Println("Видео обновлено")
	}
}

func deleteVideoFlow(videoID int) bool {
	confirm := promptui.Prompt{
		Label: "Введите 'УДАЛИТЬ' для подтверждения",
		Validate: func(input string) error {
			if input != "УДАЛИТЬ" {
				return fmt.Errorf("неверное подтверждение")
			}
			return nil
		},
	}
	_, err := confirm.Run()
	if err != nil {
		return false
	}
	err = doRequest("DELETE", fmt.Sprintf("/videos/%d", videoID), nil, nil)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return false
	}
	fmt.Println("Видео удалено")
	return true
}

func searchVideosFlow() error {
	fmt.Println("Поиск видео пока не реализован.")
	return nil
}

func listChannelVideosFlow(channelID int) error {
	page := 0
	limit := 5
	for {
		var videos []VideoResponse
		path := fmt.Sprintf("/channels/%d/videos?limit=%d&offset=%d", channelID, limit, page*limit)
		err := doRequest("GET", path, nil, &videos)
		if err != nil {
			return err
		}
		if len(videos) == 0 && page == 0 {
			fmt.Println("На этом канале пока нет видео.")
			return nil
		}
		if len(videos) == 0 {
			fmt.Println("Больше нет видео.")
			return nil
		}
		items := make([]string, len(videos)+1)
		for i, v := range videos {
			items[i] = fmt.Sprintf("%s (просмотров %d; нравится %d; не нравится %d; комментариев %d)",
				v.Title, v.Views, v.Likes, v.Dislikes, v.Comments)
		}
		items[len(videos)] = "Следующая страница →"

		selectPrompt := promptui.Select{
			Label: fmt.Sprintf("Видео канала (страница %d)", page+1),
			Items: items,
			Size:  10,
		}
		idx, _, err := selectPrompt.Run()
		if err != nil {
			return err
		}
		if idx == len(videos) {
			page++
			continue
		}
		selected := videos[idx]
		watchVideoFlow(selected.ID, selected.ChannelID)
	}
}
