package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

func ensureChannelExists() (*ChannelResponse, error) {
	var channel ChannelResponse
	err := doRequest("GET", "/channels/me", nil, &channel)
	if err == nil {
		return &channel, nil
	}
	if !isChannelNotFoundError(err) {
		return nil, err
	}
	fmt.Println("У вас ещё нет канала. Давайте создадим.")
	prompt := promptui.Prompt{Label: "Название канала"}
	name, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	prompt = promptui.Prompt{Label: "Описание (необязательно)"}
	desc, _ := prompt.Run()
	req := map[string]string{
		"channel_name": name,
		"description":  desc,
	}
	err = doRequest("POST", "/channels", req, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Канал создан!")
	err = doRequest("GET", "/channels/me", nil, &channel)
	return &channel, err
}

func isChannelNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "CHANNEL_NOT_FOUND")
}

func myChannelFlow() error {
	channel, err := ensureChannelExists()
	if err != nil {
		return err
	}
	actions := []string{
		"Просмотреть информацию",
		"Редактировать название/описание",
		"Мои видео",
		"Загрузить видео",
		"Удалить канал",
		"Назад",
	}
	for {
		selectPrompt := promptui.Select{
			Label: fmt.Sprintf("Канал: %s", channel.Name),
			Items: actions,
		}
		idx, _, err := selectPrompt.Run()
		if err != nil {
			break
		}
		switch idx {
		case 0:
			fmt.Printf("ID: %d\nНазвание: %s\nОписание: %s\nПодписчиков: %d\n",
				channel.ID, channel.Name, channel.Description, channel.SubscribersCount)
		case 1:
			editChannelFlow(channel.ID)
			var updated ChannelResponse
			if err := doRequest("GET", "/channels/me", nil, &updated); err == nil {
				channel = &updated
			}
		case 2:
			if err := myVideosFlow(); err != nil {
				fmt.Println("Ошибка:", err)
			}
		case 3:
			if err := uploadVideoFlow(); err != nil {
				fmt.Println("Ошибка:", err)
			}
		case 4:
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
			if err == nil {
				if err := doRequest("DELETE", "/channels/"+strconv.Itoa(channel.ID), nil, nil); err != nil {
					fmt.Printf("Ошибка: %v\n", err)
				} else {
					fmt.Println("Канал удалён")
					return nil
				}
			}
		case 5:
			return nil
		}
	}
	return nil
}

func editChannelFlow(channelID int) {
	prompt := promptui.Prompt{Label: "Новое название (оставьте пустым, чтобы не менять)"}
	newName, _ := prompt.Run()
	prompt = promptui.Prompt{Label: "Новое описание (оставьте пустым, чтобы не менять)"}
	newDesc, _ := prompt.Run()
	req := make(map[string]*string)
	if newName != "" {
		req["channel_name"] = &newName
	}
	if newDesc != "" {
		req["description"] = &newDesc
	}
	if len(req) == 0 {
		return
	}
	err := doRequest("PATCH", fmt.Sprintf("/channels/%d", channelID), req, nil)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Println("Канал обновлён")
	}
}
