package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"takoworld-bot/model/envconfig"
	"takoworld-bot/model/player"

	"github.com/bwmarrin/discordgo"
	"github.com/gorcon/rcon"
)

var s *discordgo.Session
var env *envconfig.Env

func init() {
	var err error

	// 環境変数のロード
	env, err = envconfig.NewEnv()
	if err != nil {
		panic(err)
	}
}

func init() {
	var err error

	// Discord のセッションを作成
	s, err = discordgo.New("Bot " + env.DISCORD_TOKEN)
	if err != nil {
		panic(err)
	}
}

var (
	integerOptionMinValue          = 1.0
	dmPermission                   = false
	defaultMemberPermissions int64 = discordgo.PermissionManageServer

	commands = []*discordgo.ApplicationCommand{
		// {
		// 	Name:        "basic-command",
		// 	Description: "Basic command",
		// },
		{
			Name:        "info",
			Description: "Show information of the TAKO WORLD Server.",
		},
		{
			Name:        "show-players",
			Description: "Show online players of the TAKO WORLD Server.",
		},
		{
			Name:        "say",
			Description: "Say everyone in the TAKO WORLD Server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "words",
					Description: "Words to say.",
					Required:    true,
				},
			},
		},
		{
			Name:        "shutdown",
			Description: "Shutdown the TAKO WORLD Server after 60s.",
		},
		{
			Name:        "show-memory",
			Description: "Show the memory usage.",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// "basic-command": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// 	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 		Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 		Data: &discordgo.InteractionResponseData{
		// 			Content: "Hey there! Congratulations, you just executed your first slash command",
		// 		},
		// 	})
		// },
		"info": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			conn, err := getRcon()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			response, err := conn.Execute("Info")
			if err != nil {
				log.Fatal(err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
		},
		"show-players": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			conn, err := getRcon()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			response, err := conn.Execute("ShowPlayers")
			if err != nil {
				log.Fatal(err)
			}

			// CSV 形式で返ってくるのでパース処理
			players, err := player.ParseCSVToPlayers(response)
			if err != nil {
				log.Fatal(err)
			}

			names := ""
			for i, p := range players {
				if i != 0 {
					names += ", "
				}
				names += p.Name
			}

			format := "%s %s %s\n"
			emoji := "🦖"
			title := fmt.Sprintf(format, emoji, "ログイン中のおたくたち", emoji)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: title,
							Color: 0x00ff00, // Green
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:  "Name",
									Value: names,
								},
							},
						},
					},
					AllowedMentions: &discordgo.MessageAllowedMentions{},
				},
			})
		},
		"say": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options

			conn, err := getRcon()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			response, err := conn.Execute("Broadcast " + options[0].StringValue())
			if err != nil {
				log.Fatal(err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
		},
		"shutdown": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			conn, err := getRcon()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			response, err := conn.Execute("Shutdown 60")
			if err != nil {
				log.Fatal(err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
		},
		"show-memory": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			response, err := exec.Command("free", "-h").Output()
			if err != nil {
				log.Fatal(err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: string(response),
				},
			})
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, env.DISCORD_GUILD_ID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Removing commands...")

	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, env.DISCORD_GUILD_ID, v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	log.Println("Gracefully shutting down.")
}

func getRcon() (*rcon.Conn, error) {
	return rcon.Dial(env.SERVER_ADDRESS+":"+env.RCON_PORT, env.RCON_PASSWORD)
}
