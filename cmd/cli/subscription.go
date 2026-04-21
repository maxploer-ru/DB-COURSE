package main

import (
	"fmt"
	"strconv"

	"github.com/manifoldco/promptui"
)

func subscribeToChannelFlow(channelID int) error {
	err := doRequest("POST", "/channels/"+strconv.Itoa(channelID)+"/subscribe", nil, nil)
	if err != nil {
		return err
	}
	fmt.Println("Вы подписались на канал.")
	return nil
}

func unsubscribeFromChannelFlow(channelID int) error {
	err := doRequest("DELETE", "/channels/"+strconv.Itoa(channelID)+"/subscribe", nil, nil)
	if err != nil {
		return err
	}
	fmt.Println("Вы отписались от канала.")
	return nil
}

func listSubscriptionsFlow() error {
	for {
		var channels []SubscriptionChannelResponse
		err := doRequest("GET", "/subscriptions", nil, &channels)
		if err != nil {
			return err
		}
		if len(channels) == 0 {
			fmt.Println("У вас нет подписок.")
			return nil
		}

		items := make([]string, len(channels))
		for i, ch := range channels {
			items[i] = fmt.Sprintf("%s (подписчиков %d)", ch.Name, ch.SubscribersCount)
		}
		items = append(items, "Вернуться в главное меню")

		prompt := promptui.Select{
			Label: "Ваши подписки",
			Items: items,
			Size:  10,
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return err
		}
		if idx == len(channels) {
			return nil
		}

		selected := channels[idx]
		fmt.Printf("\nКанал: %s\nОписание: %s\nПодписчиков: %d\nНовых видео: %d\n",
			selected.Name, selected.Description, selected.SubscribersCount, selected.NewVideosCount)

		for {
			actions := []string{
				"Просмотреть видео канала",
				"Отписаться",
				"Назад к списку подписок",
			}
			actionPrompt := promptui.Select{
				Label: "Действие",
				Items: actions,
			}
			actIdx, _, err := actionPrompt.Run()
			if err != nil {
				break
			}
			switch actIdx {
			case 0:
				if err := listChannelVideosFlow(selected.ID); err != nil {
					fmt.Printf("Ошибка: %v\n", err)
				}
			case 1:
				if err := unsubscribeFromChannelFlow(selected.ID); err != nil {
					fmt.Printf("Ошибка: %v\n", err)
				} else {
					fmt.Println("Вы отписались. Обновите список для просмотра изменений.")
				}
				break
			case 2:
				break
			}
			if actIdx == 1 || actIdx == 2 {
				break
			}
		}
	}
}

func subscriptionActionFlow(channelID int) {
	for {
		actions := []string{
			"Подписаться",
			"Отписаться",
			"Назад",
		}
		prompt := promptui.Select{
			Label: fmt.Sprintf("Управление подпиской на канал %d", channelID),
			Items: actions,
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return
		}
		switch idx {
		case 0:
			err = subscribeToChannelFlow(channelID)
		case 1:
			err = unsubscribeFromChannelFlow(channelID)
		case 2:
			return
		}
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
		} else if idx != 2 {
			fmt.Println("Готово")
		}
	}
}
