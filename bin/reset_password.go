package bin

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	apiv1 "github.com/usememos/memos/api/v1"
	apiv2 "github.com/usememos/memos/api/v2"
	"github.com/usememos/memos/internal/log"
	apiv2pb "github.com/usememos/memos/proto/gen/api/v2"
	"github.com/usememos/memos/server"
	"github.com/usememos/memos/store"
	"github.com/usememos/memos/store/db"
)

var (
	email            string
	id               int32
	password         string
	username         string
	resetPasswordCmd = &cobra.Command{
		Use:   "reset-password",
		Short: "Reset password",
		Long:  `Reset password for a supplied user id, username or email address.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			print(greetingBanner)
			fmt.Printf("\033[33mMAINTENANCE MODE: \033[1m%s\033[0m\n", cmd.Use)

			username := strings.TrimSpace(username)
			email := strings.TrimSpace(email)
			if username == "" && email == "" && id == -1 {
				fmt.Println("\033[31mERROR: user id, username or email address is required.\033[0m")
				_ = cmd.Help()
				return
			}

			if strings.TrimSpace(password) == "" {
				fmt.Println("\033[31mERROR: password can not be blank.\033[0m")
				_ = cmd.Help()
				return
			}

			ctx, cancel := context.Background(), func() {}
			dbDriver, err := db.NewDBDriver(profile)
			if err != nil {
				cancel()
				log.Error("failed to create db driver", zap.Error(err))
				return
			}
			if err := dbDriver.Migrate(ctx); err != nil {
				cancel()
				log.Error("failed to migrate db", zap.Error(err))
				return
			}

			// Port must not overlap with the main server, in case it's running.
			profile.Port -= 5

			st := store.New(dbDriver, profile)
			s, err := server.NewServer(ctx, profile, st)
			if err != nil {
				cancel()
				log.Error("failed to create server", zap.Error(err))
				return
			}

			var user apiv2pb.User

			switch {
			case id != -1:
				fmt.Printf("Resetting password for user with id \033[1m%v\033[0m\n", id)

				findUser := store.FindUser{
					ID: &id,
				}
				users, err := s.Store.ListUsers(ctx, &findUser)
				if err != nil {
					log.Error("failed to list users", zap.Error(err))
					return
				}
				if len(users) == 0 {
					fmt.Printf("\033[31mERROR: user with id \033[1m%v\033[0m not found\033[0m\n", id)
					return
				}

				username = users[0].Username
				user = apiv2pb.User{
					Name:     fmt.Sprintf("users/%s", username),
					Password: password,
				}
			case username != "":
				fmt.Printf("Resetting password for username \033[1m%v\033[0m\n", username)

				user = apiv2pb.User{
					Name:     fmt.Sprintf("users/%s", username),
					Password: strings.TrimSpace(password),
				}
			case email != "":
				fmt.Printf("Resetting password for user with email address \033[1m%v\033[0m\n", email)

				findUser := store.FindUser{
					Email: &email,
				}
				users, err := s.Store.ListUsers(ctx, &findUser)
				if err != nil {
					log.Error("failed to list users", zap.Error(err))
					return
				}
				if len(users) == 0 {
					fmt.Printf("\033[31mERROR: user with email address \033[1m%v\033[0m not found\033[0m\n", email)
					return
				}

				username = users[0].Username
				user = apiv2pb.User{
					Name:     fmt.Sprintf("users/%s", username),
					Password: strings.TrimSpace(password),
				}
			}

			// This block should be removed once the APIv1 is fully deprecated.
			// It's just checking if the password is < 3 or > 512 characters.
			{
				apiv1UpdateUserRequest := apiv1.UpdateUserRequest{
					Password: &user.Password,
				}
				validationErr := apiv1UpdateUserRequest.Validate()
				if validationErr != nil {
					fmt.Printf("\033[31mERROR: %v\033[0m\n", validationErr)
					return
				}
			}

			updateUserRequest := apiv2pb.UpdateUserRequest{
				User: &user,
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"password"},
				},
			}

			ctx = context.WithValue(ctx, apiv2.UsernameContextKey, username)
			apiV2Service := apiv2.NewAPIV2Service(s.Secret, profile, st, 0)
			updateUser, err := apiV2Service.UpdateUser(ctx, &updateUserRequest)
			if err != nil {
				cancel()
				log.Error("failed to reset password", zap.Error(err))
				return
			}

			fmt.Println(updateUser)
			fmt.Println("\033[32mSUCCESS: password reset\033[0m")
			s.Shutdown(ctx)
		},
	}
)

func init() {
	rootCmd.AddCommand(resetPasswordCmd)
	resetPasswordCmd.Flags().StringVarP(&email, "email", "", "", "Email address")
	resetPasswordCmd.Flags().Int32VarP(&id, "id", "", -1, "user id")
	resetPasswordCmd.Flags().StringVarP(&username, "username", "", "", "Username")
	resetPasswordCmd.Flags().StringVarP(&password, "password", "", "", "New password")
}
