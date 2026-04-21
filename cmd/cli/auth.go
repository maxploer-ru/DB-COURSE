package main

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

var currentUserRole string

func registerFlow() error {
	prompt := promptui.Prompt{Label: "Имя пользователя"}
	username, err := prompt.Run()
	if err != nil {
		return err
	}
	prompt = promptui.Prompt{Label: "Email"}
	email, err := prompt.Run()
	if err != nil {
		return err
	}
	prompt = promptui.Prompt{Label: "Пароль", Mask: '*'}
	password, err := prompt.Run()
	if err != nil {
		return err
	}
	req := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}
	err = doRequest("POST", "/register", req, nil)
	if err != nil {
		return err
	}
	fmt.Println("Регистрация прошла успешно! Теперь войдите в систему.")
	return nil
}

func loginFlow() error {
	prompt := promptui.Prompt{Label: "Email"}
	email, err := prompt.Run()
	if err != nil {
		return err
	}
	prompt = promptui.Prompt{Label: "Пароль", Mask: '*'}
	password, err := prompt.Run()
	if err != nil {
		return err
	}
	req := map[string]string{
		"email":    email,
		"password": password,
	}
	var resp AuthResponse
	err = doRequest("POST", "/login", req, &resp)
	if err != nil {
		return err
	}
	SetTokens(resp.AccessToken, resp.RefreshToken)
	currentUserRole = resp.User.Role
	fmt.Printf("Вы вошли как %s (%s)\n", resp.User.Username, resp.User.Email)
	return nil
}
