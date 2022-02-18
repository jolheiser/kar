package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/google/go-github/v42/github"
	"github.com/peterbourgon/ff/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

func main() {
	fs := flag.NewFlagSet("kar", flag.ExitOnError)
	token := fs.String("token", "", "GitHub token")
	kar := fs.String("kar", "", "User to (un)follow")
	interval := fs.Duration("interval", time.Minute*60*6, "Interval to (un)follow")
	debug := fs.Bool("debug", false, "Debug logging")
	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("KAR")); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *token})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	following, _, err := client.Users.IsFollowing(ctx, user.GetLogin(), *kar)
	if err != nil {
		log.Err(err).Msg("")
	}

	go func() {
		ticker := time.NewTicker(*interval)
		for {
			action := client.Users.Follow
			msg := "follow"
			if following {
				action = client.Users.Unfollow
				msg = "unfollow"
			}

			log.Debug().Msgf("%sing %s", msg, *kar)
			_, err := action(ctx, *kar)
			if err != nil {
				log.Err(err).Msgf("could not %s %s", msg, *kar)
			}

			following = !following
			<-ticker.C
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Kill, os.Interrupt)
	<-ch
}
