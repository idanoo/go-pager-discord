package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/joho/godotenv"
)

var (
	DISCORD_WEBHOOK string
	RTL_FREQ        string
	RTL_DEVICE_ID   string
	RTL_GAIN        string

	discordClient webhook.Client
	socketFile    = "/var/run/pager.sock"

	// Emojies to replace
	emojis = map[string][]string{
		":house:":                     []string{"house", "roof"},
		":blue_car:":                  []string{"mva", "mvc", "car "},
		":fire:":                      []string{"fire", "smoke", "smoking", "chimney"},
		":ocean:":                     []string{"flood"},
		":evergreen_tree:":            []string{"tree"},
		":wind_blowing_face:":         []string{"wind", "blow"},
		":zap:":                       []string{"power", "electric", "spark"},
		":ambulance: :purple_circle:": []string{"purple"},
		":ambulance:":                 []string{"chest pain", "not alert", "seizure", "breath", "pain", "fall ", "orange", "sweats", "acute", "red 1", "red 2", "respitory", "bleeding", "purple"},
		":biohazard":                  []string{"hazchem"},
		":helmet_with_cross:":         []string{"resc "},
	}

	// Phrases to skip
	skip = []string{
		"this is a test",
		"enabled demodulators",
		"test page",
		"assigned to station",
	}
)

// Load .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Webhook to post to
	DISCORD_WEBHOOK = os.Getenv("DISCORD_WEBHOOK")
	if DISCORD_WEBHOOK == "" {
		log.Fatal("DISCORD_WEBHOOK empty or invalid")
	}

	// Frequency to listen to
	RTL_FREQ = os.Getenv("RTL_FREQ")
	if RTL_FREQ == "" {
		log.Fatal("RTL_FREQ empty or invalid")
	}

	// Not required - if not set, skip setting
	RTL_DEVICE_ID = os.Getenv("RTL_DEVICE_ID")

	// Receiver gain
	RTL_GAIN = os.Getenv("RTL_GAIN")
	if RTL_FREQ == "" {
		log.Fatal("RTL_GAIN empty or invalid")
	}
}

// Start func
func main() {
	log.Println("Booting go-pager-discord")

	// Load discord client
	client, err := webhook.NewWithURL(DISCORD_WEBHOOK)
	if err != nil {
		log.Fatal(err)
	}
	discordClient = client

	// Start running rtl_fm + multimon
	deviceID := ""
	if RTL_DEVICE_ID != "" {
		deviceID = fmt.Sprintf("-d %s", RTL_DEVICE_ID)
	}

	// If the socket doesn't exist.. run command!
	if _, err := os.Stat(socketFile); !errors.Is(err, os.ErrNotExist) {

	}

	// Build command
	script := fmt.Sprintf(
		"rtl_fm -M fm -d %s -f %sM -g %s -s 22050 -- | multimon-ng -t raw -a POCSAG512 -a POCSAG1200 -a FLEX -a POCSAG2400 /var/run/pager.sock",
		deviceID,
		RTL_FREQ,
		RTL_GAIN,
	)
	cmd := exec.Command(script)
	cmd.Stdout = os.Stdout
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	// Wait for it to load
	time.Sleep(1 * time.Second)

	// Create a Unix domain socket and listen for incoming connections.
	socket, err := net.Listen("unix", "/tmp/echo.sock")
	if err != nil {
		log.Fatal(err)
	}

	// Cleanup the sockfile.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove("/tmp/echo.sock")
		os.Exit(1)
	}()

	for {
		// Accept an incoming connection.
		conn, err := socket.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// Handle the connection in a separate goroutine.
		go func(conn net.Conn) {
			defer conn.Close()
			// Create a buffer for incoming data.
			buf := make([]byte, 4096)

			// Read data from the connection.
			n, err := conn.Read(buf)
			if err != nil {
				log.Fatal(err)
			}

			// Echo the data back to the connection.
			_, err = conn.Write(buf[:n])
			if err != nil {
				log.Fatal(err)
			}
		}(conn)
	}
}

func sendMessage(msg string) {
	// Send to discord!
	_, err := discordClient.CreateMessage(discord.WebhookMessageCreate{
		Content: msg,
	})

	if err != nil {
		log.Print(err)
	}
}
