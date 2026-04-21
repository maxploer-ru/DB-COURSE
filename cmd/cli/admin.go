package main

import (
	"fmt"
	"strconv"

	"github.com/manifoldco/promptui"
)

func adminMenu() {
	for {
		options := []string{
			"Забанить пользователя",
			"Разбанить пользователя",
			"Назад",
		}
		prompt := promptui.Select{
			Label: "Админ-меню",
			Items: options,
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return
		}
		switch idx {
		case 0:
			banUserFlow()
		case 1:
			unbanUserFlow()
		case 2:
			return
		}
	}
}

func banUserFlow() {
	prompt := promptui.Prompt{
		Label: "Введите ID пользователя для бана",
		Validate: func(input string) error {
			_, err := strconv.Atoi(input)
			if err != nil {
				return fmt.Errorf("введите число")
			}
			return nil
		},
	}
	idStr, err := prompt.Run()
	if err != nil {
		return
	}
	userID, _ := strconv.Atoi(idStr)

	err = banUser(userID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Println("✅ Пользователь забанен.")
	}
}

func unbanUserFlow() {
	prompt := promptui.Prompt{
		Label: "Введите ID пользователя для разбана",
		Validate: func(input string) error {
			_, err := strconv.Atoi(input)
			if err != nil {
				return fmt.Errorf("введите число")
			}
			return nil
		},
	}
	idStr, err := prompt.Run()
	if err != nil {
		return
	}
	userID, _ := strconv.Atoi(idStr)

	err = unbanUser(userID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Println("✅ Пользователь разбанен.")
	}
}
