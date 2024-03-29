package main

import (
	_ "embed"
	"fmt"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"log"
	"math/rand"
	"strings"
	"time"
)

type Anna struct {
	ircevent.Connection
	channels map[string]string
	words    map[string]bool
	colours  []string
}

func must(err error) {
	if err != nil {
		log.Fatalf("Error invoking IRC command: %v", err)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	bot := &Anna{
		Connection: ircevent.Connection{
			Server:        "irc:6667",
			Nick:          "Anna",
			RealName:      "Anna",
			UseSASL:       false,
			UseTLS:        false,
			Debug:         true,
			ReconnectFreq: time.Second,
		},
		channels: make(map[string]string),
		colours:  strings.Split(colours, "\n"),
		words:    loadWords(),
	}

	// Oper up as soon as we connect
	bot.AddConnectCallback(func(message ircmsg.Message) {
		must(bot.Send("OPER", "anna", "*Qc$ZRDB8WT2Kw"))
	})

	// Once we're an oper, join the admin channel and set it up
	bot.AddCallback("381", func(message ircmsg.Message) {
		must(bot.Send("SAJOIN", bot.CurrentNick(), "#admin"))
		must(bot.Send("SAMODE", "#admin", "+Cnsti"))
	})

	// Listen for server notices that tell us about clients connecting etc
	bot.AddCallback("NOTICE", func(message ircmsg.Message) {
		if !strings.ContainsRune(message.Source, '.') {
			// Ignore notices that aren't from servers
			return
		}

		// See if it's a client disconnecting - if so we'll close the channel created for them
		var nick string
		n, _ := fmt.Sscanf(message.Params[1], "\x0314-\x0fQUIT\x0314-\x03 %s exited the network", &nick)
		if n == 1 {
			nick = strings.TrimSuffix(nick, "\x0F")
			must(bot.Send("PRIVMSG", "#admin", fmt.Sprintf("User disconnected - %s - closed channel %s", nick, bot.channels[nick])))
			must(bot.Send("PART", bot.channels[nick]))
			delete(bot.channels, nick)
		}

		// See if it's a client connecting - if so we'll make a channel for them and SAJOIN them into it
		n, _ = fmt.Sscanf(message.Params[1], "\x0314-\x0fCONNECT\x0314-\x03 Client connected [%s", &nick)
		if n == 1 {
			nick = strings.TrimSuffix(nick, "]")
			log.Printf("User connected: %s", nick)
			channel := randChannel()
			bot.channels[nick] = channel
			must(bot.Send("SAJOIN", bot.CurrentNick(), channel))
			must(bot.Send("TOPIC", channel, "Welcome! Feel free to chat but please be sure to obey all channel rules."))
			must(bot.Send("SAJOIN", nick, channel))
			must(bot.Send("PRIVMSG", "#admin", fmt.Sprintf("New user connected - %s - joined to channel %s", nick, channel)))
		}

		// See if it's a nick change - so we can update our tracking of which channel belongs to which user
		var newNick string
		n, _ = fmt.Sscanf(message.Params[1], "\x0314-\x0fNICK\x0314-\x03 %s changed nickname to %s", &nick, &newNick)
		if n == 2 {
			nick = strings.TrimSuffix(nick, "\x0F")
			log.Printf("User changed names: %s -> %s", nick, newNick)
			must(bot.Send("PRIVMSG", "#admin", fmt.Sprintf("User changed nicks - %s -> %s", nick, newNick)))
			bot.channels[newNick] = bot.channels[nick]
			delete(bot.channels, nick)
		}
	})

	// Listen for normal messages, check they comply with the rules
	bot.AddCallback("PRIVMSG", func(message ircmsg.Message) {
		body := strings.Join(message.Params[1:], " ")
		nick := strings.Split(message.Source, "!")[0]
		channel := message.Params[0]

		result := bot.checkMessage(body)
		must(bot.Send("PRIVMSG", "#admin", fmt.Sprintf("User %s said '%s' in %s result: %s", nick, body, channel, result)))
		if result != correctString {
			must(bot.Send("KICK", channel, nick, result))
		}
	})

	must(bot.Connection.Connect())

	bot.Loop()
}

//go:embed words.txt
var wordList string

//go:embed colours.txt
var colours string

func loadWords() map[string]bool {
	res := make(map[string]bool)
	words := strings.Split(wordList, "\n")
	for i := range words {
		res[words[i]] = true
	}
	return res
}

const correctString = "Correct"

// checkMessage checks if the given text complies with our rules. Returns either the rule that was violated (in a
// format suitable for passing to a user) or - if no rules were broken - correctString.
func (a *Anna) checkMessage(text string) string {
	text = strings.ToLower(text)
	for i := range text {
		if text[i] < 'a' || text[i] > 'z' {
			return "Rule 1: Messages must not contain spaces, punctuation or special characters."
		}
	}

	if len(text) < 4 {
		return "Rule 2: Messages must be at least 4 letters long."
	}

	if len(text) > 10 {
		return "Rule 3: Ain't nobody got time to read messages longer than 10 letters."
	}

	if !a.words[text] {
		return "Rule 4: Non-dictionary words are forbidden."
	}

	found := ""
	for i := range a.colours {
		if strings.Contains(text, a.colours[i]) {
			found = a.colours[i]
			break
		}
	}
	if found == "" {
		return "Rule 5: Messages must have a colour from Joseph's Technicolour Dreamcoat somewhere within them."
	}

	if len(text)%2 != 1 {
		return "Rule 6: Messages must be an odd number of letters long."
	}

	if len(text) != 9 {
		return "Rule 7: Messages must have a non-prime length."
	}

	if !strings.Contains(text, "peach") {
		return "Rule 8: The colour from rule 6 must be peach."
	}

	if text != "impeached" {
		return "Rule 9: Messages must contain a word that describes a public official accused of a crime."
	}

	return correctString
}

// randChannel returns a random, 32-character long, channel name
func randChannel() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 32)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return "#" + string(b)
}
