/*
Copyright 2020 Matt Hamilton

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/mxschmitt/playwright-go"
	"github.com/spf13/viper"
)

// Alpine should be set to true with ldflags during build
// if the image is Alpine Linux. This _should_ be a boolean,
// but must be a string to add at compile-time.
var Alpine string

var conf *config

// config holds our program's internal configuration. See config.sample.yml
type config struct {
	AlURL           string `mapstructure:"al_url"`
	AlUsername      string `mapstructure:"al_username"`
	AlPassword      string `mapstructure:"al_password"`
	AlCharacterName string `mapstructure:"al_character_name"`
}

func main() {
	conf = getConfig()
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("unable to launch playwright: %s\n", err.Error())
	}
	opts := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
			"--no-zygote",
			"--single-process",
			"--disable-default-apps",
			"--disable-filesystem",
		},
		ChromiumSandbox: playwright.Bool(false),
	}
	if Alpine != "" {
		opts.ExecutablePath = playwright.String("/usr/bin/chromium-browser")
	}
	browser, err := pw.Chromium.Launch(opts)
	if err != nil {
		log.Fatalf("unable to lauch Chromium: %s\n", err.Error())
	}
	browserCtx, err := browser.NewContext()
	if err != nil {
		log.Fatalf("unable to get Chromium context: %s\n", err.Error())
	}
	err = authToAdventureLand(browserCtx)
	if err != nil {
		log.Fatalf("failed to authenticate to Adventure Land: %s\n", err.Error())
	}
	character, err := NewCharacter(browserCtx, conf.AlCharacterName)
	if err != nil {
		log.Fatalf("unable to create character: %s\n", err.Error())
	}

	// create an asynchronous chracter consumer
	var wg sync.WaitGroup
	wg.Add(1)
	ch := make(chan *Character, 1) // we buffer one update
	go func() {
		for { // forever
			char, ok := <-ch // pull a character from the channel
			if !ok {
				wg.Done() // signal program exit from within goroutine
				return
			}
			if err := char.Update(); err != nil {
				log.Fatal(fmt.Errorf("unable to update character: %w\n", err))
			}
			// our char object has now been synced with the game state
			//TODO: do stuff here!
			log.Printf("character: %s | Level: %d | HP: %d/%d\n",
				char.Name,
				char.Level,
				char.HP,
				char.MaxHP,
			)
			time.Sleep(1 * time.Second) // we update once a second
			ch <- char                  // add the char back to the top of the stack
		}
	}()

	// add our character
	ch <- character

	// and wait until an error or exit
	wg.Wait()
}

// Character is deserialized from the "character" object in the headless browser
type Character struct {
	Name  string `json:"name" mapstructure:"name"`
	HP    int64  `json:"hp" mapstructure:"hp"`
	MaxHP int64  `json:"max_hp" mapstructure:"max_hp"`
	Level int64  `json:"level" mapstructure:"level"`

	page *playwright.Page // the browser page
}

// Update is called to pull the latest state from the browser
func (c *Character) Update() error {
	resp, err := c.page.Evaluate("character")
	if err != nil {
		return err
	}
	characterObject, ok := resp.(map[string]interface{})
	if !ok {
		return fmt.Errorf("\"character\" object did not match an expected type")
	}
	if err := mapstructure.Decode(characterObject, c); err != nil {
		return fmt.Errorf("error unmarshaling character JSON to struct: %w", err)
	}
	return nil
}

// NewCharacter spawns a new tab for each character.
func NewCharacter(browserCtx *playwright.BrowserContext, characterName string) (*Character, error) {
	page, err := browserCtx.NewPage()
	if err != nil {
		return nil, err
	}
	time.Sleep(4 * time.Second)
	_, err = page.Goto(
		"https://adventure.land/character/"+characterName+"/in/US/III/?no_graphics=true",
		playwright.PageGotoOptions{
			WaitUntil: playwright.String("load"),
		},
	)
	if err != nil {
		return nil, err
	}
	time.Sleep(5 * time.Second)
	return &Character{
		Name: characterName,
		page: page,
	}, nil
}

// authToAdventureLand uses playwright to interact with the page.
func authToAdventureLand(browserCtx *playwright.BrowserContext) error {
	//TODO: replace the sleeps with proper checks ;-)
	page, err := browserCtx.NewPage()
	if err != nil {
		return fmt.Errorf("unable to create new page in Chromium: %w", err)
	}
	_, err = page.Goto(conf.AlURL, playwright.PageGotoOptions{
		WaitUntil: playwright.String("load"),
	})
	if err != nil {
		return fmt.Errorf("error navigating to Adventure Land site: %w", err)
	}
	time.Sleep(5 * time.Second)
	_, err = page.Evaluate("$('#loginbuttons').hide(); $('#loginlogin').show(); on_resize()")
	if err != nil {
		return fmt.Errorf("error showing login window on Adventure Land site: %w", err)
	}
	time.Sleep(2 * time.Second)
	_ = page.Fill("#email2", conf.AlUsername)
	_ = page.Fill("#password2", conf.AlPassword)
	time.Sleep(1 * time.Second)
	_, err = page.Evaluate("api_call_l('signup_or_login',{email:$('#email2').val(),password:$('#password2').val(),only_login:true},{disable:$(this)})")
	if err != nil {
		return fmt.Errorf("error signing into Adventure Land site: %w", err)
	}
	// close the page after a delay
	go func(p *playwright.Page) {
		//NOTE: This is a good place to debug the code (this is just after auth)
		time.Sleep(1 * time.Second)
		p.Close() // the characters each create their own page. Safe to close this page post-auth.
	}(page)
	return nil
}

// getConfig reads in our program options
func getConfig() *config {
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./configs")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	conf := &config{}
	if err := viper.Unmarshal(conf); err != nil {
		log.Fatal(err)
	}
	return conf
}
