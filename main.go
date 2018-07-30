package main

import (
	"flag"
	"strings"

	"github.com/ovh/go-ovh/ovh"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/TheoBrigitte/ovh-dyndns/pkg/lookup/hostname"
	"github.com/TheoBrigitte/ovh-dyndns/pkg/lookup/ip"
	"github.com/TheoBrigitte/ovh-dyndns/pkg/ovh/client"
)

const (
	configFileName = "ovh-dyndns"
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
		flag.String("record-type", "", "DNS record typ (e.g. CNAME)")
		flag.String("subdomains", "", "DNS sub domain filter. Multiple entries separated by comma (e.g. www,ftp)")

		pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
		pflag.Parse()
		viper.BindPFlags(pflag.CommandLine)

		// Load config file
		viper.SetConfigName(configFileName)
		viper.AddConfigPath("/etc/")
		viper.AddConfigPath("$HOME/")
		viper.AddConfigPath(".")
		err = viper.ReadInConfig()
		if err != nil {
			log.Fatal(err)
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
			log.Fatal(err)
			return
		}

		ovhClient = client.New(*c)
	}

	if generateConsumerKey {
		// Generate customer key.
		fields := log.Fields{
			"status": "generating",
		}
		log.WithFields(fields).Info("customer key")
		res, err := ovhClient.GenerateConsumerKey()
		if err != nil {
			log.Fatal(err)
			return
		}
		fields["status"] = "done"
		log.WithFields(fields).Info("customer key")

		log.Printf("response\n%v", res)
		return
	}

	// Read public IP adresse.
	var ipAdresse string
	{
		fields := log.Fields{
			"status": "reading",
		}
		log.WithFields(fields).Info("ip adresse")
		ipAdresse, err = ip.GetPublic()
		if err != nil {
			log.Fatal(err)
			return
		}
		fields["status"] = "done"
		fields["ip"] = ipAdresse
		log.WithFields(fields).Info("ip adresse")
	}

	// Lookup hostname.
	var hostnames []string
	{
		if recordType == "CNAME" {
			fields := log.Fields{
				"status": "reading",
			}
			log.WithFields(fields).Info("hostname")
			hostnames, err = hostname.Get(ipAdresse)
			if err != nil {
				log.Fatal(err)
				return
			}
			fields["status"] = "done"
			fields["hostname"] = hostnames[0]
			log.WithFields(fields).Info("hostname")
		}
	}

	// Update dns records.
	for _, domain := range subDomains {
		// Find the dns record.
		fields := log.Fields{
			"status": "finding",
			"domain": domain,
		}
		log.WithFields(fields).Info("dns record")
		id, err := ovhClient.FindRecord(zone, recordType, domain)
		if err != nil {
			log.Fatal(err)
			return
		}
		fields["record ID"] = id

		// Read the dns record
		fields["status"] = "reading"
		log.WithFields(fields).Info("dns record")
		record, err := ovhClient.GetRecord(zone, id)
		if err != nil {
			log.Fatal(err)
			return
		}
		fields["record"] = record
		log.WithFields(fields).Info("dns record")

		// Update the dns record
		fields["status"] = "updating"
		switch record.FieldType {
		case "CNAME":
			record.Target = hostnames[0]
		case "A":
			record.Target = ipAdresse
		default:
			log.Fatalf("unsuported record type: %q", record.FieldType)
			return
		}

		err = ovhClient.UpdateRecord(zone, id, *record)
		if err != nil {
			log.Fatal(err)
			return
		}
		fields["status"] = "done"
		log.WithFields(fields).Info("dns record")
	}

	log.Info("success")
}