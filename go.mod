module github.com/ljanyst/peroxide

go 1.15

// These dependencies are `replace`d below, so the version numbers should be ignored.
// They are in a separate require block to highlight this.
require (
	github.com/emersion/go-imap v1.0.6
	github.com/emersion/go-message v0.12.1-0.20201221184100-40c3f864532b
	github.com/jameskeane/bcrypt v0.0.0-20170924085257-7509ea014998 // indirect
)

require (
	github.com/ProtonMail/go-crypto v0.0.0-20211221144345-a4f6767435ab
	github.com/ProtonMail/go-rfc5322 v0.8.0
	github.com/ProtonMail/go-srp v0.0.1
	github.com/ProtonMail/go-vcard v0.0.0-20180326232728-33aaa0a0c8a5
	github.com/ProtonMail/gopenpgp/v2 v2.4.1
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/emersion/go-imap-appendlimit v0.0.0-20190308131241-25671c986a6a
	github.com/emersion/go-imap-move v0.0.0-20190710073258-6e5a51a5b342
	github.com/emersion/go-imap-quota v0.0.0-20210203125329-619074823f3c
	github.com/emersion/go-imap-unselect v0.0.0-20171113212723-b985794e5f26
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21
	github.com/emersion/go-smtp v0.14.0
	github.com/emersion/go-textwrapper v0.0.0-20200911093747-65d896831594
	github.com/emersion/go-vcard v0.0.0-20190105225839-8856043f13c5 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-resty/resty/v2 v2.6.0
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/jaytaylor/html2text v0.0.0-20200412013138-3577fbdbcff7
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-isatty v0.0.14
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/miekg/dns v1.1.41
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/ricochet2200/go-disk-usage/du v0.0.0-20210707232629-ac9918953285
	github.com/sirupsen/logrus v1.7.0
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stretchr/testify v1.7.0
	github.com/vmihailenco/msgpack/v5 v5.1.3
	go.etcd.io/bbolt v1.3.6
	golang.org/x/crypto v0.0.0-20220507011949-2cf3adece122
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	golang.org/x/sys v0.0.0-20220111092808-5a964db01320 // indirect
	golang.org/x/text v0.3.7
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace (
	github.com/emersion/go-imap => github.com/ProtonMail/go-imap v0.0.0-20201228133358-4db68cea0cac
	github.com/emersion/go-message => github.com/ProtonMail/go-message v0.0.0-20210611055058-fabeff2ec753
	github.com/jameskeane/bcrypt => github.com/ProtonMail/bcrypt v0.0.0-20210511135022-227b4adcab57
)
