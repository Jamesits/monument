# monument

Allow a file to be decrypted when and only when you die.

## Project Status

Under heavy development, no backward compatibility is guaranteed. Please use the same version for encryption & decryption.

## Design

A successful decryption of a monument-encrypted file requires all the following conditions to be true:

* The encrypted version of your secret file is accessible
* Your death, incapacitation or any trigger that you need to set up beforehand has been triggered (A Dead Man's Switch, referred as DMS later)
* Keys from _k_ out of the _n_ key holders are gathered

At encryption, monument generates a new PGP keypair, encrypts your file and splits the private key using Shamir's Secret Sharing Algorithm. You can designate how many people you will give one key to, and how many people are required to finally decrypt your file. You will receive _n_ keys, one for every person, plus _m_ keys to put into your DMS service.

After encryption, you need to set up your DMS service in a way that it will send all the _m_ keys to every of the _n_ people, tell them about your death and tell them how to contact each other. 

When you die and the DMS service successfully triggers, if _k_ out of the _n_ people managed to have contact, they will have _k + m_ keys in total which will allow monument to finally decrypt your secret file.

## Usage

### Encryption phase

For example, if you have 5 people to give keys to, and 3 of them are required to decrypt the secret: 

```shell
monument encrypt --name "Your Legal Name" --email "your.email@example.com" --people 5 --decryptable 3 --file secret-message.txt --output out
```

All the files required for decryption will be put into `out` directory. You need to:

* Publish `public/secret-message.txt.gpg` to a place where it will be available even if you die
* Put the content of `secret/shares_for_death_switch.txt` to your DMS service
* Hand out the keys in `secret/shares_for_people.txt`, one key per person
* Delete your original `secret-message.txt` file and all the keys in `secret/*`

### Decryption phase

First download the encrypted `secret-message.txt.gpg`. Then run monument to start the decryption phase:

```shell
monument decrypt --file secret-message.txt.gpg
```

Monument will then ask you for the keys you gathered. Paste one key per line. If the keys are correct, the secret will be revealed.