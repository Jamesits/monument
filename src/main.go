package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/SSSaaS/sssa-golang"
	"golang.org/x/crypto/openpgp"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	if len(os.Args) < 2 {
		subCommandDoNotMatch()
		os.Exit(1)
	}

	switch strings.ToLower(os.Args[1]) {
	case "encrypt":
		doEncryption()
	case "decrypt":
		doDecryption()
	default:
		subCommandDoNotMatch()
		os.Exit(1)
	}
}

func printHelpHeader() {
	fmt.Fprintf(os.Stderr, "monument %s \nhttps://github.com/Jamesits/monument\n\n", getVersionNumberString())
	// fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
}

func subCommandDoNotMatch() {
	printHelpHeader()
	fmt.Fprintf(os.Stderr, "Available subcommands: \n\tencrypt\n\tdecrypt\n")
}

func doEncryption() {
	var err error
	// parse command line
	encryptionCmdLineFlags := flag.NewFlagSet("encrypt", flag.ExitOnError)
	name := encryptionCmdLineFlags.String("name", "", "Your legal name")
	email := encryptionCmdLineFlags.String("email", "", "Your email")
	encryptionFilePath := encryptionCmdLineFlags.String("file", "-", "Path to the file you want to encrypt")
	totalPeople := encryptionCmdLineFlags.Int("people", 0, "How many people in total do you need to give part of the key")
	requiredPeople := encryptionCmdLineFlags.Int("decryptable", 0, "How many people is required to decrypt the file")
	outputDir := encryptionCmdLineFlags.String("output", "output", "A directory where all files you need will be put")

	err = encryptionCmdLineFlags.Parse(os.Args[2:])
	hardFailIf(err)

	// TODO: check arguments

	// create output directory
	publicDirPath := path.Join(*outputDir, "public")
	secretDirPath := path.Join(*outputDir, "secret")
	err = os.MkdirAll(publicDirPath, 0700)
	hardFailIf(err)
	err = os.MkdirAll(secretDirPath, 0700)
	hardFailIf(err)

	// init monument internal data
	m := monument{
		ownerName:     *name,
		ownerEmail:    *email,
		totalPeople:   *totalPeople,
		minimalPeople: *requiredPeople,
	}

	initMonument(&m)

	log.Printf("\nWill generate %d keys, %d for distribution and %d for death switch\nAllow decryption at %d keys or %d people\n", m.totalShares, m.totalPeople, m.deadSwitchShares, m.minimalShares, m.minimalPeople)

	// encrypt file
	// TODO: support stdin
	encryptionFileReader, err := os.Open(*encryptionFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer encryptionFileReader.Close()

	encryptionFileWriter, err := os.Create(path.Join(publicDirPath, filepath.Base(*encryptionFilePath)+".gpg"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer encryptionFileWriter.Close()

	err = encryptData([]*openpgp.Entity{m.pgpEntity}, m.pgpEntity, encryptionFileReader, encryptionFileWriter)
	hardFailIf(err)

	// output pubkey
	publicKeyWriter, err := os.Create(path.Join(publicDirPath, "pubkey.gpg"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer publicKeyWriter.Close()
	exportPublicKey(&m, publicKeyWriter)

	// output shares for the death switch
	sharesForDeathSwitch, err := os.Create(path.Join(secretDirPath, "shares_for_death_switch.txt"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sharesForDeathSwitch.Close()

	_, err = fmt.Fprintf(sharesForDeathSwitch, "Put all the following lines into a \"dead man's switch\" service: \n\n")
	hardFailIf(err)
	for _, share := range m.shamirShares[0 : m.deadSwitchShares-1] {
		_, err = fmt.Fprintln(sharesForDeathSwitch, share)
		hardFailIf(err)
	}

	// output shares for people
	sharesForPeopleWriter, err := os.Create(path.Join(secretDirPath, "shares_for_people.txt"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sharesForPeopleWriter.Close()

	_, err = fmt.Fprintf(sharesForPeopleWriter, "Send each line to a different people: \n\n")
	hardFailIf(err)
	for _, share := range m.shamirShares[m.deadSwitchShares:] {
		_, err = fmt.Fprintln(sharesForPeopleWriter, share)
		hardFailIf(err)
	}
}

func doDecryption() {
	var err error
	// parse command line
	decryptionCmdLineFlags := flag.NewFlagSet("decrypt", flag.ExitOnError)
	encryptedFilePath := decryptionCmdLineFlags.String("file", "-", "Path to the file you want to decrypt")
	privateKeyPath := decryptionCmdLineFlags.String("private-key", "", "Use this provate key instead of reading from shares")

	err = decryptionCmdLineFlags.Parse(os.Args[2:])
	hardFailIf(err)

	// open encrypted file
	encryptedFileReader, err := os.Open(*encryptedFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer encryptedFileReader.Close()

	var privateKeyReader io.Reader
	var privateKeyBuffer string

	if len(*privateKeyPath) > 0 {
		// read private key
		privateKeyFileReader, err := os.Open(*privateKeyPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer privateKeyFileReader.Close()
		privateKeyReader = privateKeyFileReader
	} else {
		// read shares
		var collectedShares []string

		// read shamir shares
		// TODO: support reading from a file or something
		fmt.Println("Please paste the keys, one key per line:")

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			// TODO: check if line is empty
			collectedShares = append(collectedShares, scanner.Text())
			privateKeyBuffer, err = shamirDecrypt(collectedShares)

			if errors.Is(err, sssa.ErrOneOfTheSharesIsInvalid) {
				// invalid share
				_, err = fmt.Fprintln(os.Stderr, "Last key is invalid")
				hardFailIf(err)
				collectedShares = collectedShares[0 : len(collectedShares)-2]
			}

			if strings.Contains(privateKeyBuffer, "BEGIN PGP PRIVATE KEY BLOCK") {
				// yes we decrypted the block
				// fmt.Println(privateKeyBuffer)
				break
			}
		}

		if err := scanner.Err(); err != nil {
			log.Println(err)
		}

		privateKeyReader = strings.NewReader(privateKeyBuffer)
	}

	// https://gist.github.com/stuart-warren/93750a142d3de4e8fdd2
	// TODO: check if privateKeyBuffer is valid
	var entityList openpgp.EntityList
	entityList, err = openpgp.ReadArmoredKeyRing(privateKeyReader)
	hardFailIf(err)

	// assert len(entityList) > 0
	// fmt.Printf("Entities: %d\n", len(entityList))

	for key, _ := range entityList[0].Identities {
		// here key == value.name
		fmt.Printf("Decrypted identity: %s\n", key)
	}

	// Decrypt it with the contents of the private key
	md, err := openpgp.ReadMessage(encryptedFileReader, entityList, nil, nil)
	hardFailIf(err)
	bytes, err := ioutil.ReadAll(md.UnverifiedBody)
	hardFailIf(err)
	fmt.Println("Contents: ")
	fmt.Println(string(bytes))
}
