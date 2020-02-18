package main

import (
	"bytes"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	_ "golang.org/x/crypto/ripemd160"
	"io"
	"strconv"
	"time"
)

const (
	md5       = 1
	sha1      = 2
	ripemd160 = 3
	sha256    = 8
	sha384    = 9
	sha512    = 10
	sha224    = 11
)

/*
Minimal operation to initialize a new monument:

1. generate an OpenPGP keypair
2. save the public key
3. distribute the private key via Shamir's Secret Sharing algorithm
 */

func initMonument(monument *monument) {
	monument.createdAt = time.Now()

	// generate an OpenPGP keypair
	var err error
	monument.pgpEntity, err = openpgp.NewEntity(monument.ownerName, "", monument.ownerEmail, nil)
	hardFailIf(err)

	// Sign all the identities

	// dur := uint32(config.Expiry.Seconds())
	for _, id := range monument.pgpEntity.Identities {
		// id.SelfSignature.KeyLifetimeSecs = &dur

		id.SelfSignature.PreferredSymmetric = []uint8{
			uint8(packet.CipherAES256),
			uint8(packet.CipherAES192),
			uint8(packet.CipherAES128),
			uint8(packet.CipherCAST5),
			uint8(packet.Cipher3DES),
		}

		id.SelfSignature.PreferredHash = []uint8{
			sha256,
			sha1,
			sha384,
			sha512,
			sha224,
		}

		id.SelfSignature.PreferredCompression = []uint8{
			uint8(packet.CompressionZLIB),
			uint8(packet.CompressionZIP),
		}

		err = id.SelfSignature.SignUserId(id.UserId.Id, monument.pgpEntity.PrimaryKey, monument.pgpEntity.PrivateKey, nil)
		hardFailIf(err)
	}

	// Self-sign the Subkeys
	// https://github.com/alokmenghrajani/gpgeez/blob/master/gpgeez.go
	for _, subkey := range monument.pgpEntity.Subkeys {
		// subkey.Sig.KeyLifetimeSecs
		err := subkey.Sig.SignKey(subkey.PublicKey, monument.pgpEntity.PrivateKey, nil)
		hardFailIf(err)
	}

	//w, err := armor.Encode(os.Stdout, openpgp.PublicKeyType, nil)
	//hardFailIf(err)
	//defer w.Close()
	//
	//monument.pgpEntity.Serialize(w)

	// serialize the private key
	var privateKeyBuffer bytes.Buffer
	w, err := armor.Encode(&privateKeyBuffer, openpgp.PrivateKeyType, generateKeyBlockHeader(monument))
	hardFailIf(err)
	err = monument.pgpEntity.SerializePrivate(w, nil)
	hardFailIf(err)
	err = w.Close()
	hardFailIf(err)

	// create Shamir shares of the serialized private key
	monument.totalShares, monument.minimalShares, monument.deadSwitchShares = getPeopleCount(monument.totalPeople, monument.minimalPeople)
	monument.shamirShares = shamirEncrypt(privateKeyBuffer.String(), monument.totalShares, monument.minimalShares)
}

func generateKeyBlockHeader(monument *monument) (ret map[string]string) {
	ret = map[string]string{
		"CreatedBy": getVersionFullString(),
		"CreatedAt": monument.createdAt.String() + " (" + strconv.FormatInt(monument.createdAt.UTC().UnixNano(), 10) + ")",
	}

	return
}

// note: RIPEMD160 is used by default
// https://github.com/golang/go/issues/12153
func encryptData(recipients []*openpgp.Entity, signer *openpgp.Entity, r io.Reader, w io.Writer) error {
	wc, err := openpgp.Encrypt(w, recipients, signer, &openpgp.FileHints{IsBinary: true}, nil)
	if err != nil {
		return err
	}
	if _, err := io.Copy(wc, r); err != nil {
		return err
	}
	return wc.Close()
}

func exportPublicKey(monument *monument, writer io.Writer) {
	var err error
	armoredWriter, err := armor.Encode(writer, openpgp.PublicKeyType, generateKeyBlockHeader(monument))
	hardFailIf(err)
	defer armoredWriter.Close()

	err = monument.pgpEntity.Serialize(armoredWriter)
	hardFailIf(err)
}

func decryptData() {

}