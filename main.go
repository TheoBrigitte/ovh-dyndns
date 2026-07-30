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
			log.Fatalf("viper.BindPFlags %v", err)
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
			log.Fatalf("viper.ReadInConfig %v", err)
			return
		}

		zone = viper.GetString("zone")
		recordType = strings.ToUpper(viper.GetString("record-type"))
		generateConsumerKey = viper.GetBool("consumer-key")
		subDomains = strings.Split(viper.GetString("subdomains"), ",")

		log.WithFields(log.Fields{
			"zone":                zone,
			"record-type":         recordType,
			"subdomains":          subDomains,
			"generateConsumerKey": generateConsumerKey,
		}).Info("arguments")
	}

	// Create client for OVH API.
	var ovhClient client.Client
	{
		// Create a client using credentials from config files or environment variables
		c, err := ovh.NewDefaultClient()
		if err != nil {
			log.Fatalf("ovh.NewDefaultClient %v", err)
			return
		}

		ovhClient = client.New(*c)
	}

	if generateConsumerKey {
		// Generate consumer key.
		fields := log.Fields{
			"status": "generating",
		}
		log.WithFields(fields).Info("consumer key")
		res, err := ovhClient.GenerateConsumerKey()
		if err != nil {
			log.Fatalf("ovhClient.GenerateConsumerKey %v", err)
			return
		}
		fields["status"] = statusDone
		log.WithFields(fields).Info("consumer key")

		log.Printf("response\n%v", res)
		return
	}

	// Read public IP address.
	var ipAddress string
	{
		fields := log.Fields{
			"status": "reading",
		}
		log.WithFields(fields).Info("ip address")
		ipAddress, err = ip.GetPublic()
		if err != nil {
			log.Fatalf("ip.GetPublic %v", err)
			return
		}
		fields["status"] = statusDone
		fields["ip"] = ipAddress
		log.WithFields(fields).Info("ip address")
	}

	// Lookup hostname.
	var hostnames []string
	{
		if recordType == "CNAME" {
			fields := log.Fields{
				"status": "reading",
			}
			log.WithFields(fields).Info("hostname")
			hostnames, err = hostname.Get(ipAddress)
			if err != nil {
				log.Fatalf("hostname.Get %v", err)
				return
			}
			fields["status"] = statusDone
			fields["hostname"] = hostnames[0]
			log.WithFields(fields).Info("hostname")
		}
	}

	// Update dns records.
	for _, domain := range subDomains {
		// Find the dns record.
		fields := log.Fields{
			"status":     "finding",
			"domain":     domain,
			"zone":       zone,
			"recordType": recordType,
		}
		log.WithFields(fields).Info("dns record")
		id, err := ovhClient.FindRecord(zone, recordType, domain)
		if err != nil {
			if !errors.Is(err, ovherrors.RecordNotFound) {
				log.Fatalf("ovhClient.FindRecord %v", err)
				return
			}
		}
		if id != 0 {
			fields["ID"] = id
			fields["status"] = "reading"
			log.WithFields(fields).Info("dns record")

			// Read the dns record
			record, err := ovhClient.GetRecord(zone, id)
			if err != nil {
				log.Fatalf("ovhClient.GetRecord %v", err)
				return
			}

			fields["record"] = record
			fields["status"] = "updating"
			log.WithFields(fields).Info("dns record")

			// Update the dns record
			var target string
			switch record.FieldType {
			case "CNAME":
				// TODO: fix potential nil pointer dereference if hostnames is empty
				target = hostnames[0]
			case "A":
				target = ipAddress
			default:
				log.Fatalf("unsupported record type: %q", record.FieldType)
				return
			}
			// Skip update when target is the same.
			if record.Target == target {
				fields["status"] = "ok"
				log.WithFields(fields).Info("dns record")
				continue
			}

			record.Target = target
			err = ovhClient.UpdateRecord(zone, id, *record)
			if err != nil {
				log.Fatalf("ovhClient.UpdateRecord %v", err)
				return
			}
			fields["status"] = "updated"
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
			fields["status"] = "creating"
			log.WithFields(fields).Info("dns record")
			newRecord, err := ovhClient.CreateRecord(zone, record)
			if err != nil {
				log.Fatalf("ovhClient.CreateRecord %v", err)
				return
			}
			fields["record"] = newRecord
			fields["status"] = "created"
		}

		log.WithFields(fields).Info("dns record")
	}

	log.Info("success")
}
