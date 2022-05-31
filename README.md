
peroxide
========

⚠ **Warning**: This software has not been thoroughly reviewed for security.
You should only use it if you know what you're doing. I strongly advise against
running it on the open Internet.

Peroxide is a fork of the [ProtonMail bridge][1]. Its goal is to be much like
[Hydroxide][2] except with as much re-use of the upstream code as possible. The
re-use ensures that the upstream changes to the service APIs can be merged in as
fast and as efficiently as possible. At the same time, Peroxide aims to run as a
server providing data access using standard protocols so that a wide variety of
devices can use their native productivity tools.

To that end, Peroxide:

 * is buildable using plain `go build`
 * drops the original GUI and CLI
 * drops all the desktop integration and trackers
 * drops dependence on binary packages
 * drops the integrated upgrade functionality
 * unables multiple device-specific passwords for every account
 * encrypts the ProtonMail credentials on disk and does not require any external
   secret store to do that
 * user-supplied passwords are keys used to decrypt the credentials in memory; they
   are never stored on disk

Server setup
------------

⚠ **Warning**: This software has not been thoroughly reviewed for security.
You should only use it if you know what you're doing. I strongly advise against
running it on the open Internet.

Run the `install.sh` script to install peroxide in your system.

Peroxide reads its settings from a configuration file located in
`/etc/peroxide.conf` by default. This configuration file holds a bunch of
key-value pairs in YAML format. There's an example in the root of the source
tree in a file called `config.example.yaml`.

The package provides two executables:

 * `peroxide` - the program that interacts with ProtonMail's services and acts
   as an IMAP and SMTP server for the email clients
 * `peroxide-cfg` - the program that manages the user accounts, login keys, and
   implements other helper functions

Peroxide encrypts the IMAP and SMTP communication with the clients using TLS and
will not work without a valid certificate. You can either use a service like
Let's Encrypt to get a certificate signed by a trusted CA or use `peroxide-cfg`
to generate a self-signed one. Running:

    ]==> sudo -u peroxide peroxide-cfg -action gen-x509 -x509-org "my-organization" -x509-cn "my-hostname"

will generate `cert.pem` and `key.pem` files in the current working directory.
These files must be copied to the location where the server expects them, as
configured in `peroxide.conf`. By default, it's: `/etc/peroxide/`.

You can then enable the service by typing:

    ]==> sudo systemctl enable peroxide
    ]==> sudo systemctl start peroxide

User management
---------------

To log in to your ProtonMail account, type:

    ]==> sudo -u peroxide peroxide-cfg -action login-account -account-name foo

It will authenticate you with the ProtonMail's services and print a
random-generated key. Please note this key; it will be needed to add
device-specific keys or re-login.

To add a device-specific key type:

    ]==> sudo -u peroxide peroxide-cfg -action add-key -account-name foo -key-name test

The command will add a device-specific key called `test` to the user account
`foo` and print that key to standard output. As above, this key is not stored
anywhere, but it must be used for authentication in your email program.

For the settings described above, the emain client configuration would be:

 * **Login:** `foo..test@protonmail.com` (appending `..test` to the username
   portion of the login selects the device-specific key named `test`)
 * **Password:** The random key printed by the configuration program when adding
   the device-specific key
 * **SMTP/IMAP server:** The address of the server running peroxide
 * **SMTP Port:** 1025
 * **IMAP Port:** 1143
 * **Encryption:** STARTTLS for both SMTP and IMAP

`peroxide-cfg` provides a bunch of other functions dealing with user and key
management described in the program's help message. Any change to the
configuration, including adding accounts or keys, necessitates a restart of the
server.

Device Configuration
--------------------

When working with laptops or desktop computers, it's easy to enter this
configuration data by hand into whatever program you need. The
`cmd/mobileconfig-gen` directory contains a program that generates device
configuration files for iOS. It takes JSON as input:


    ]==> ./mobileconfig-gen -in account.json -out account.mobileconfig

You can upload this file to some secret location (it contains your passwords)
and generate the QR code pointing to it like this:

    ]==> qrencode -t ansiutf8 https://secret.location/of/the/mobile/config/file

Then, scan this code with your device's camera.

[1]: https://github.com/ProtonMail/proton-bridge
[2]: https://github.com/emersion/hydroxide
