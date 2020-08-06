package updater

// TODO: Regarding the DatabaseMirror setting: Maybe KubeAV could host its own, cluster-internal mirror?
const freshclamTemplate = `
DatabaseDirectory /var/lib/clamav
LogTime yes
LogVerbose no
Checks 12
Foreground yes

DatabaseMirror database.clamav.net

# Don't do this; this is very memory intensive
TestDatabases no
`
