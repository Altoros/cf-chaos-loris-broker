package main

import (
	"fmt"
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/Altoros/cf-chaos-loris-broker/broker"
	chaos_loris_cleint "github.com/Altoros/cf-chaos-loris-broker/client"
	"github.com/Altoros/cf-chaos-loris-broker/cmd"
	"github.com/Altoros/cf-chaos-loris-broker/config"
	database "github.com/Altoros/cf-chaos-loris-broker/db"
	goflags "github.com/jessevdk/go-flags"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pivotal-cf/brokerapi"
)

func main() {
	brokerLogger := lager.NewLogger("service-broker")
	brokerLogger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	brokerLogger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	opts := cmd.CommandOpts{}
	_, err := goflags.ParseArgs(&opts, os.Args[1:])

	brokerLogger.Info("Using config file: " + opts.ConfigPath)

	config, err := config.LoadFromFile(opts.ConfigPath)
	if err != nil {
		brokerLogger.Error("Failed to load the config file", err, lager.Data{
			"broker-config-path": opts.ConfigPath,
		})
		return
	}

	var db *gorm.DB
	db, err = database.New()
	if err != nil {
		brokerLogger.Error("Failed to connect to the mysql: %s", err)
	}
	defer db.Close()

	client := chaos_loris_cleint.New(opts.ChaosLorisHost)

	serviceBroker := broker.NewServiceBroker(
		client,
		opts,
		config,
		db,
		brokerLogger,
	)

	credentials := brokerapi.BrokerCredentials{
		Username: opts.ChaosLorisUsername,
		Password: opts.ChaosLorisPassword,
	}

	brokerAPI := brokerapi.New(serviceBroker, brokerLogger, credentials)

	http.Handle("/", brokerAPI)
	brokerLogger.Info("Listening for requests", lager.Data{
		"port": opts.Port,
	})

	err = http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), nil)
	if err != nil {
		brokerLogger.Error("Failed to start the server", err)
	}
}
