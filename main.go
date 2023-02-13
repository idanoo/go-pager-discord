package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	DISCORD_WEBHOOK string
	RTL_FREQ        string
	RTL_DEVICE_ID   string
	RTL_GAIN        string

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

	// Start running rtl_fm + multimon
	deviceID := ""
	if RTL_DEVICE_ID != "" {
		deviceID = fmt.Sprintf("-d %s", RTL_DEVICE_ID)
	}

	// Build command
	_ = fmt.Sprintf(
		"rtl_fm -M fm -d %s -f %sM -g %s -s 22050 -- | multimon-ng -t raw -a POCSAG512 -a POCSAG1200 -a FLEX -a POCSAG2400 /dev/stdin",
		deviceID,
		RTL_FREQ,
		RTL_GAIN,
	)
}
