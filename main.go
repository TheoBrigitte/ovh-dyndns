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

	// recordsKey is the configuration file key holding the records table.
	recordsKey = "records"
)

// Record is a single entry of the records table in the configuration file.
type Record struct {
	Zone       string   `mapstructure:"zone"`
	RecordType string   `mapstructure:"record-type"`
	Subdomains []string `mapstructure:"subdomains"`
}

func main() {
	var err error

	// Define and process arguments.
	flag.Bool("consumer-key", false, "generate a consumer key")
	flag.Bool("version", false, "print version and exit")

	// Parse flags and bind them to viper
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatalf("Initializing flags %v", err)
		return
	}

	// Print version and exit if the version flag is set.
	if viper.GetBool("version") {
		if version.Version == "" {
			version.Version = "dev" // Set a default version if not provided
		}
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

	var records []Record
	err = viper.UnmarshalKey(recordsKey, &records)
	if err != nil {
		log.Fatalf("Parsing %s from configuration file %v", recordsKey, err)
		return
	}

	run(records)
}

func run(records []Record) {
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

	if viper.GetBool("consumer-key") {
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

	if len(records) == 0 {
		log.Fatalf("No %s found in configuration file %s", recordsKey, viper.ConfigFileUsed())
		return
	}

	log.WithFields(log.Fields{"records": len(records)}).Info("Starting ovh-dyndns")

	var l lookup
	for _, r := range records {
		recordType := strings.ToUpper(r.RecordType)
		fields := log.Fields{
			"zone":        r.Zone,
			"record-type": recordType,
			"subdomains":  r.Subdomains,
		}

		// Resolve the target once per entry, all its sub domains share it.
		target, err := l.target(recordType)
		if err != nil {
			log.WithFields(fields).Errorf("Resolving DNS record target %v", err)
			continue
		}

		for _, subDomain := range r.Subdomains {
			err = updateRecord(ovhClient, r.Zone, recordType, subDomain, target)
			if err != nil {
				log.WithFields(fields).Errorf("%v", err)
			}
		}
	}
}

// lookup resolves DNS record targets and caches them, so each lookup is
// performed at most once per run.
type lookup struct {
	ipv4     string
	ipv6     string
	hostname string
}

// target returns the value to be set on a DNS record of the given type.
func (l *lookup) target(recordType string) (string, error) {
	switch recordType {
	case "A":
		if l.ipv4 == "" {
			log.Info("Reading public IPv4 address")
			ipv4, err := ip.GetPublicIPv4()
			if err != nil {
				return "", fmt.Errorf("reading public IPv4 address: %w", err)
			}
			log.WithFields(log.Fields{"ip": ipv4}).Info("Found public IPv4 address")
			l.ipv4 = ipv4
		}

		return l.ipv4, nil
	case "AAAA":
		if l.ipv6 == "" {
			log.Info("Reading public IPv6 address")
			ipv6, err := ip.GetPublicIPv6()
			if err != nil {
				return "", fmt.Errorf("reading public IPv6 address: %w", err)
			}
			log.WithFields(log.Fields{"ip": ipv6}).Info("Found public IPv6 address")
			l.ipv6 = ipv6
		}

		return l.ipv6, nil
	case "CNAME":
		if l.hostname == "" {
			// Hostname is the reverse lookup of the public IPv4 address.
			ipv4, err := l.target("A")
			if err != nil {
				return "", err
			}

			log.Info("Reading public hostname")
			hostnames, err := hostname.Get(ipv4)
			if err != nil {
				return "", fmt.Errorf("reading public hostname: %w", err)
			}
			log.WithFields(log.Fields{"hostname": hostnames[0]}).Info("Found public hostname")
			l.hostname = hostnames[0]
		}

		return l.hostname, nil
	default:
		return "", fmt.Errorf("unsupported record type %q", recordType)
	}
}

// updateRecord creates or updates a single DNS record so it points to target,
// then refreshes the DNS zone.
func updateRecord(c client.Client, zone, recordType, subDomain, target string) error {
	fields := log.Fields{
		"zone":        zone,
		"record-type": recordType,
		"subdomain":   subDomain,
	}

	// Find the dns record.
	log.WithFields(fields).Info("Searching DNS record")
	id, err := c.FindRecord(zone, recordType, subDomain)
	if err != nil {
		if !errors.Is(err, ovherrors.RecordNotFound) {
			return fmt.Errorf("searching DNS record: %w", err)
		}
		log.WithFields(fields).Info("DNS record not found")
	}

	if id == 0 {
		// Create the dns record.
		record := client.DNSRecord{
			FieldType: recordType,
			SubDomain: subDomain,
			Target:    target,
		}

		fields["record"] = record
		log.WithFields(fields).Info("Creating DNS record")
		newRecord, err := c.CreateRecord(zone, record)
		if err != nil {
			return fmt.Errorf("creating DNS record: %w", err)
		}
		fields["record"] = newRecord
		log.WithFields(fields).Info("Successfully created DNS record")
	} else {
		// Read the dns record.
		record, err := c.GetRecord(zone, id)
		if err != nil {
			return fmt.Errorf("reading DNS record: %w", err)
		}
		fields["ID"] = id
		fields["record"] = record
		log.WithFields(fields).Info("Found DNS record")

		// Skip update when target is the same.
		if record.Target == target {
			log.WithFields(fields).Info("Skipping DNS record, already up to date")
			return nil
		}

		record.Target = target

		log.WithFields(fields).Info("Updating DNS record")
		err = c.UpdateRecord(zone, id, *record)
		if err != nil {
			return fmt.Errorf("updating DNS record: %w", err)
		}
		log.WithFields(fields).Info("Successfully updated DNS record")
	}

	err = c.RefreshZone(zone)
	if err != nil {
		return fmt.Errorf("refreshing DNS zone: %w", err)
	}

	return nil
}
