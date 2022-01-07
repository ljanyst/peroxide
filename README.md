
peroxide
========

Peroxide is a fork of the [ProtonMail bridge][1]. Its goal is to be much like
[Hydroxide][2] except with as much re-use of the upstream code as possible. The
reason for the re-use is to ensure that the changes to the service APIs can be
merged in as fast and as easily as possible. At the same time, Peroxide aims to:

 * run as a server providing data access using standard protocols, so that a
   wide variety of devices can use their native productivity tools instead of
   ProtonMail's proprietary ones

 * implement features that are missing from the upstream version because they
   are hard to make work with Outlook

 * make things easy to hack on without a deluge of dependencies providing little
   value in the context of the two above points

To than end, Perixide:

 * is buildable with `go build`
 * drops the original GUI and CLI
 * drops all the desktop desktop integration and trackers
 * provides a server program and a separate configuration program

1: https://github.com/ProtonMail/proton-bridge
2: https://github.com/emersion/hydroxide
