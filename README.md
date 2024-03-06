## Mail Downloader
### Download mail from mail server from cli

```
Usage:
  mail_downloader [flags]

Flags:
  -P, --ImapPort int               imap ImapPort (default 993)
      --emailFolders stringArray   email folders (default [INBOX,SENT])
      --endTime string             email end time,layout:2006-01-02 15:04:05
  -h, --help                       help for mail_downloader
      --maxUID uint32              email max uid
      --minUID uint32              email min uid
  -p, --password string            email password
  -o, --savePath string            email save path (default "./emails")
  -s, --server string              imap server
  -c, --size int                   size of fetch email onetime,size > 0 and size<=50,size should not be too large when most of emails have attachments (default 5)
      --startTime string           email start time,layout:2006-01-02 15:04:05
  -u, --username string            email username
