package main

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

func runMainMenu() {
	for {
		if !IsLoggedIn() {
			authMenu()
			continue
		}
		options := []string{
			"Все видео",
			"Поиск видео",
			"Мой канал",
			"Подписки",
			"Выйти",
			"Выход",
		}
		prompt := promptui.Select{
			Label: "Главное меню",
			Items: options,
			Size:  10,
		}
		idx, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Ошибка ввода: %v\n", err)
			return
		}
		switch idx {
		case 0:
			if err := listAllVideosFlow(); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case 1:
			if err := searchVideosFlow(); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case 2:
			if err := myChannelFlow(); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case 3:
			if err := listSubscriptionsFlow(); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case 4:
			ClearTokens()
			fmt.Println("Вы вышли из системы.")
		case 5:
			fmt.Println("До свидания!")
			os.Exit(0)
		}
	}
}

func authMenu() {
	prompt := promptui.Select{
		Label: "Добро пожаловать в ZVideo TUI",
		Items: []string{"Войти", "Регистрация", "Выход"},
	}
	idx, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Ошибка ввода: %v\n", err)
		os.Exit(1)
	}
	switch idx {
	case 0:
		if err := loginFlow(); err != nil {
			fmt.Printf("Ошибка входа: %v\n", err)
		}
	case 1:
		if err := registerFlow(); err != nil {
			fmt.Printf("Ошибка регистрации: %v\n", err)
		}
	case 2:
		fmt.Println("До свидания!")
		os.Exit(0)
	}
}
