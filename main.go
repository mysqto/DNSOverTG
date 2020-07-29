// Copyright (c) 2019, Chen Lei <my@mysq.to>
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mysqto/dig"
	"github.com/mysqto/log"

	tgBot "gopkg.in/tucnak/telebot.v2"
)

// TelegramBotAPIBase is the API base of telegram Bot API.
const TelegramBotAPIBase = "https://api.telegram.org"

var queries, success uint64
var userVisited big.Int
var users uint64

func visited(userId int) {
	if userVisited.Bit(userId) == 0 {
		atomic.AddUint64(&users, 1)
		userVisited.SetBit(&userVisited, userId, 1)
	}
}

func main() {
	flag.Parse()

	bot, err := tgBot.NewBot(tgBot.Settings{
		Token:  os.Getenv("BOT_KEY"),
		URL:    TelegramBotAPIBase,
		Poller: &tgBot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatalf("error creating Telegram Bot: %v", err)
		return
	}

	bot.Handle(tgBot.OnText, func(m *tgBot.Message) {

		answer, err := dig.Dig(strings.Fields(m.Text))

		visited(m.Sender.ID)

		atomic.AddUint64(&queries, 1)

		if err != nil {
			_, _ = bot.Send(m.Sender, fmt.Sprintf("error querying with : %s, %v", m.Text, err))
		} else {
			atomic.AddUint64(&success, 1)
		}

		log.Infof("user: %s, query=[%s], err=[%v]", m.Sender.Username, m.Text, err)

		_, _ = bot.Send(m.Sender, answer)
	})

	bot.Handle(tgBot.OnPhoto, func(m *tgBot.Message) {
		// photos only
	})

	bot.Handle(tgBot.OnChannelPost, func(m *tgBot.Message) {
		// channel posts only
	})

	// Command: /start
	bot.Handle("/start", func(m *tgBot.Message) {
		if !m.Private() {
			return
		}

		_, _ = bot.Send(m.Sender, `DNS Over Telegram, fully support unix dig command, brought to you  by @mysqto, try send @8.8.8.8 t.me to me`)
	})

	bot.Handle("/status", func(m *tgBot.Message) {
		if !m.Private() {
			return
		}

		_, _ = bot.Send(m.Sender, fmt.Sprintf("users: %v, total : %v, success: %v", users, queries, success))
	})

	bot.Handle("/help", func(m *tgBot.Message) {
		if !m.Private() {
			return
		}
		answer, _ := dig.Dig(strings.Fields("-h"))
		_, _ = bot.Send(m.Sender, answer)
	})

	bot.Start()
}
