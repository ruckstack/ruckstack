package ops

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"golang.org/x/crypto/bcrypt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

func ListUsers() error {
	ctx := context.Background()
	client := kube.Client()

	secret, err := client.CoreV1().Secrets("ops").Get(ctx, "ops-users", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error accessing user list: %s", err)
	}

	ui.Println("Existing Users")
	ui.Println("---------------------------------")

	if len(secret.Data) == 0 {
		ui.Println("    None")
	}

	users := []string{}
	for user, _ := range secret.Data {
		users = append(users, user)
	}
	sort.Strings(users)

	for _, user := range users {
		ui.Println(user)
	}

	return nil
}

func AddUser() error {
	ui.Println("Add user to /ops")
	ui.Println("---------------------------------")

	username := ui.PromptForString("Username", "", ui.NotEmptyCheck)
	password := ui.PromptForPassword("Password", ui.NotEmptyCheck)
	_ = ui.PromptForPassword("Re-enter Password", func(input string) error {
		if input != password {
			return fmt.Errorf("passwords do not match")
		}
		return nil
	})

	ctx := context.Background()
	client := kube.Client()

	secrets := client.CoreV1().Secrets("ops")
	secret, err := secrets.Get(ctx, "ops-users", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error accessing user list: %s", err)
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	_, alreadyExists := secret.Data[username]
	if alreadyExists {
		return fmt.Errorf("Username %s already exists", username)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	secret.Data[username] = hashedPassword

	_, err = secrets.Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error saving new user: %s", err)
	}

	return nil
}

func DeleteUser() error {
	ui.Println("Delete user from /ops")
	ui.Println("---------------------------------")

	username := ui.PromptForString("Username to remove", "", ui.NotEmptyCheck)

	ctx := context.Background()
	client := kube.Client()

	secrets := client.CoreV1().Secrets("ops")
	secret, err := secrets.Get(ctx, "ops-users", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error accessing user list: %s", err)
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	_, alreadyExists := secret.Data[username]
	if !alreadyExists {
		return fmt.Errorf("Unknown user: %s", username)
	}

	if !ui.PromptForBoolean(fmt.Sprintf("Delete %s", username), &ui.FalseBoolean) {
		ui.Println("\nUser not removed")
		return nil
	}

	delete(secret.Data, username)

	_, err = secrets.Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error saving new user: %s", err)
	}

	ui.Println("\nUser removed")

	return nil
}
