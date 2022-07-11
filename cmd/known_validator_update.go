package cmd

import (
	"github.com/flashbots/boost-relay/beaconclient"
	"github.com/flashbots/boost-relay/common"
	"github.com/flashbots/boost-relay/datastore"
	"github.com/flashbots/go-boost-utils/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	valUpdateDefaultBeaconURI = "http://localhost:3500"
)

func init() {
	rootCmd.AddCommand(knownValidatorUpdateCmd)
	knownValidatorUpdateCmd.Flags().StringVar(&beaconNodeURI, "beacon-uri", valUpdateDefaultBeaconURI, "beacon endpoint")
	knownValidatorUpdateCmd.Flags().StringVar(&redisURI, "redis-uri", defaultredisURI, "redis uri")

	knownValidatorUpdateCmd.Flags().BoolVar(&logJSON, "json", defaultLogJSON, "log in JSON format instead of text")
	knownValidatorUpdateCmd.Flags().StringVar(&logLevel, "loglevel", defaultLogLevel, "log-level: trace, debug, info, warn/warning, error, fatal, panic")
	knownValidatorUpdateCmd.Flags().SortFlags = false
}

var knownValidatorUpdateCmd = &cobra.Command{
	Use:   "known-validator-update",
	Short: "Update the known validators in Redis",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		common.LogSetup(logJSON, logLevel)
		log := logrus.WithField("module", "cmd/known-validator-update")
		log.Infof("boost-relay %s", version)

		// Connect beacon client to node
		if beaconNodeURI == "" {
			log.Fatal("beacon-uri is required")
		}
		log.Infof("Using beacon endpoint: %s", beaconNodeURI)
		beaconClient := beaconclient.NewProdBeaconClient(log, beaconNodeURI)

		// Check beacon node status
		_, err = beaconClient.SyncStatus()
		if err != nil {
			log.WithError(err).Fatal("error checking beacon-node sync status")
		}

		// Connect to Redis
		if redisURI == "" {
			log.Fatal("redis-uri is required")
		}
		redis, err := datastore.NewRedisCache(redisURI)
		if err != nil {
			log.WithError(err).Fatalf("Failed to connect to Redis at %s", redisURI)
		}
		log.Infof("Connected to Redis at %s", redisURI)

		// Query beacon node for known validators
		log.Info("Querying validators from beacon node... (this may take a while)")
		validators, err := beaconClient.FetchValidators()
		if err != nil {
			log.WithError(err).Fatal("failed to fetch validators from beacon node")
		}
		log.Infof("Got %d validators from BN", len(validators))

		// Update Redis with validators
		log.Info("Writing to Redis...")

		// redis.SetKnownValidators(validators)
		for _, v := range validators {
			err = redis.SetKnownValidator(types.PubkeyHex(v.Validator.Pubkey))
			if err != nil {
				log.WithError(err).WithField("pubkey", v.Validator.Pubkey).Fatal("failed to set known validator in Redis")
			}
		}
		log.Info("Updated Redis")
	},
}
