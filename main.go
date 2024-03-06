package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// 后续有新增情况，建议配置到apollo
const (
	InboxFolder   = "INBOX"
	SentFolder    = "SENT"
	TimeLayoutStr = "2006-01-02 15:04:05"
)

var cmd = &cobra.Command{
	Use:          "mail_downloader",
	RunE:         cmdF,
	SilenceUsage: true,
}

func main() {
	cmd.Flags().StringP("server", "s", "", "imap server")
	cmd.Flags().IntP("ImapPort", "P", 993, "imap ImapPort")
	cmd.Flags().StringP("username", "u", "", "email username")
	cmd.Flags().StringP("password", "p", "", "email password")
	cmd.Flags().StringP("startTime", "", "", "email start time,layout:"+TimeLayoutStr)
	cmd.Flags().StringP("endTime", "", "", "email end time,layout:"+TimeLayoutStr)
	cmd.Flags().Uint32P("minUID", "", 0, "email min uid")
	cmd.Flags().Uint32P("maxUID", "", 0, "email max uid")
	cmd.Flags().IntP("size", "c", 5, "size of fetch email onetime,size > 0 and size<=50,size should not be too large when most of emails have attachments")
	cmd.Flags().StringArrayP("emailFolders", "", []string{InboxFolder, SentFolder}, "email folders")
	cmd.Flags().StringP("savePath", "o", "./emails", "email save path")

	cmd.SetArgs(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func cmdF(command *cobra.Command, args []string) error {
	imapHost, err := command.Flags().GetString("server")
	if err != nil {
		return err
	}
	imapPort, err := command.Flags().GetInt("ImapPort")
	if err != nil {
		return err
	}
	username, err := command.Flags().GetString("username")
	if err != nil {
		return err
	}
	password, err := command.Flags().GetString("password")
	if err != nil {
		return err
	}

	if len(imapHost) == 0 || len(username) == 0 || len(password) == 0 {
		return errors.New("email server,username,password are required")
	}

	startTime, err := command.Flags().GetString("startTime")
	if err != nil {
		return err
	}
	endTime, err := command.Flags().GetString("endTime")
	if err != nil {
		return err
	}
	minUID, err := command.Flags().GetUint32("minUID")
	if err != nil {
		return err
	}
	maxUID, err := command.Flags().GetUint32("maxUID")
	if err != nil {
		return err
	}
	batchSize, err := command.Flags().GetInt("size")
	if err != nil {
		return err
	}
	if batchSize > 50 {
		batchSize = 50
	}
	if batchSize <= 0 {
		batchSize = 5
	}

	emailFolders, err := command.Flags().GetStringArray("emailFolders")
	if err != nil {
		return err
	}
	savePath, err := command.Flags().GetString("savePath")
	if err != nil {
		return err
	}
	emailSrv := Service{
		OutputDirectory: savePath,
		ImapPort:        imapPort,
		ImapHost:        imapHost,
		Account: AccountInfo{
			Username: username,
			Password: password,
		},
	}
	for _, folder := range emailFolders {
		if err1 := emailSrv.AutoFetch(context.Background(), folder, startTime, endTime, minUID, maxUID, batchSize); err != nil {
			log.Printf("FetchEmail.AutoFetch:error:%v\n", err1)
			return err1
		}
	}
	return nil
}
