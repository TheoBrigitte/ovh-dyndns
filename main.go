package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/ovh/go-ovh/ovh"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/TheoBrigitte/ovh-dyndns/pkg/lookup/hostname"
	"github.com/TheoBrigitte/ovh-dyndns/pkg/lookup/ip"
	"github.com/TheoBrigitte/ovh-dyndns/pkg/ovh/client"
	ovherrors "github.com/TheoBrigitte/ovh-dyndns/pkg/ovh/errors"
)

const (
	// name is the name of the application.
	name = "ovh-dyndns"

	configFileName = "ovh-dyndns"
	statusDone     = "done"
)

func main() {
	var (
		zone                string
		recordType          string
		generateConsumerKey bool
		subDomains          []string

		err error
	)
	{
		// Define and process arguments.
		flag.Bool("consumer-key", false, "generate a consumer key")
		flag.String("zone", "", "DNS zone name (e.g. example.com)")
		flag.String("record-type", "", "DNS record type (e.g. CNAME)")
		flag.String("subdomains", "", "DNS sub domain filter. Multiple entries separated by comma (e.g. www,ftp)")
		flag.Bool("version", false, "print version and exit")

		pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
		pflag.Parse()
		err = viper.BindPFlags(pflag.CommandLine)
		if err != nil {
			log.Fatalf("Initializing flags %v", err)
			return
		}

		if viper.GetBool("version") {
			if version.Version == "" {
				version.Version = "dev" // Set a default version if not provided
			}
			// If the version flag is set, print the version and exit.
			fmt.Println(version.Print(name))
			return
		}

		// Load config file
		viper.SetConfigName(configFileName)
		viper.AddConfigPath("/etc/ovh-dyndns")
		viper.AddConfigPath("$HOME/")
		viper.AddConfigPath(".")
		err = viper.ReadInConfig()
		if err != nil {
			log.Fatalf("Reading configuration file %v", err)
			return
		}

		zone = viper.GetString("zone")
		recordType = strings.ToUpper(viper.GetString("record-type"))
		generateConsumerKey = viper.GetBool("consumer-key")
		subDomains = strings.Split(viper.GetString("subdomains"), ",")

		log.WithFields(log.Fields{
			"zone":        zone,
			"record-type": recordType,
			"subdomains":  subDomains,
		}).Info("Starting ovh-dyndns")
	}

	// Create client for OVH API.
	var ovhClient client.Client
	{
		// Create a client using credentials from config files or environment variables
		c, err := ovh.NewDefaultClient()
		if err != nil {
			log.Fatalf("Initializing OVH client %v", err)
			return
		}

		ovhClient = client.New(*c)
	}

	if generateConsumerKey {
		// Generate consumer key.
		log.Info("Generating consumer key")
		res, err := ovhClient.GenerateConsumerKey()
		if err != nil {
			log.Fatalf("Generating consumer key %v", err)
			return
		}
		log.Info("Successfully generated consumer key")
		log.Printf("response\n%v", res)
		return
	}

	// Read public IP address.
	var ipAddress string
	{
		log.Info("Reading public IP address")
		ipAddress, err = ip.GetPublic()
		if err != nil {
			log.Fatalf("Reading public IP address %v", err)
			return
		}
		fields := log.Fields{"ip": ipAddress}
		log.WithFields(fields).Info("Found public IP address")
	}

	// Lookup hostname.
	var hostnames []string
	{
		if recordType == "CNAME" {
			log.Info("Reading public Hostname")
			hostnames, err = hostname.Get(ipAddress)
			if err != nil {
				log.Fatalf("Reading public Hostname %v", err)
				return
			}
			fields := log.Fields{"hostnames": hostnames[0]}
			log.WithFields(fields).Info("Found public Hostname")
		}
	}

	// Update dns records.
	for _, domain := range subDomains {
		// Find the dns record.
		fields := log.Fields{
			"domain":     domain,
			"zone":       zone,
			"recordType": recordType,
		}
		log.WithFields(fields).Info("Searching DNS record")
		id, err := ovhClient.FindRecord(zone, recordType, domain)
		if err != nil {
			if !errors.Is(err, ovherrors.RecordNotFound) {
				log.WithFields(fields).Errorf("Searching DNS record %v", err)
				continue
			}
			log.WithFields(fields).Info("DNS record not found")
		}
		if id != 0 {
			// Read the dns record
			record, err := ovhClient.GetRecord(zone, id)
			if err != nil {
				log.Errorf("Reading DNS record %v", err)
				continue
			}
			fields["ID"] = id
			fields["record"] = record
			log.WithFields(fields).Info("Found DNS record")

			// Update the dns record
			var target string
			switch record.FieldType {
			case "CNAME":
				// TODO: fix potential nil pointer dereference if hostnames is empty
				target = hostnames[0]
			case "A":
				target = ipAddress
			default:
				log.Errorf("Unsupported record type: %q", record.FieldType)
				continue
			}
			// Skip update when target is the same.
			if record.Target == target {
				log.WithFields(fields).Info("Skipping DNS record, already up to date")
				continue
			}

			record.Target = target

			log.WithFields(fields).Info("Updating DNS record")

			err = ovhClient.UpdateRecord(zone, id, *record)
			if err != nil {
				log.Errorf("Updating DNS record %v", err)
				continue
			}
			log.WithFields(fields).Info("Successfully updated DNS record")
		} else {
			// Create the dns record
			record := client.DNSRecord{
				FieldType: recordType,
				SubDomain: domain,
			}
			if recordType == "CNAME" {
				record.Target = hostnames[0]
			} else {
				record.Target = ipAddress
			}

			fields["record"] = record
			log.WithFields(fields).Info("Creating DNS record")
			newRecord, err := ovhClient.CreateRecord(zone, record)
			if err != nil {
				log.Errorf("Creating DNS record %v", err)
				continue
			}
			fields["record"] = newRecord
			log.WithFields(fields).Info("Successfully created DNS record")
		}
	}
}
