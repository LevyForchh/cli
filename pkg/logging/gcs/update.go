package gcs

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to update GCS logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	Bucket            common.OptionalString
	User              common.OptionalString
	SecretKey         common.OptionalString
	Path              common.OptionalString
	Period            common.OptionalUint
	FormatVersion     common.OptionalUint
	GzipLevel         common.OptionalUint8
	Format            common.OptionalString
	ResponseCondition common.OptionalString
	TimestampFormat   common.OptionalString
	MessageType       common.OptionalString
	Placement         common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a GCS logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the GCS logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the GCS logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("bucket", "The bucket of the GCS bucket").Action(c.Bucket.Set).StringVar(&c.Bucket.Value)
	c.CmdClause.Flag("user", "Your GCS service account email address. The client_email field in your service account authentication JSON").Action(c.User.Set).StringVar(&c.User.Value)
	c.CmdClause.Flag("secret-key", "Your GCS account secret key. The private_key field in your service account authentication JSON").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("path", "The path to upload logs to (default '/')").Action(c.Path.Set).StringVar(&c.Path.Value)
	c.CmdClause.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").Action(c.Period.Set).UintVar(&c.Period.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").Action(c.GzipLevel.Set).Uint8Var(&c.GzipLevel.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).Action(c.TimestampFormat.Set).StringVar(&c.TimestampFormat.Value)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.MessageType.Set).StringVar(&c.MessageType.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateGCSInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	gcs, err := c.Globals.Client.GetGCS(&fastly.GetGCSInput{
		Service: serviceID,
		Name:    c.EndpointName,
		Version: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateGCSInput{
		Service:           gcs.ServiceID,
		Version:           gcs.Version,
		Name:              gcs.Name,
		NewName:           gcs.Name,
		Bucket:            gcs.Bucket,
		User:              gcs.User,
		SecretKey:         gcs.SecretKey,
		Path:              gcs.Path,
		Period:            gcs.Period,
		FormatVersion:     gcs.FormatVersion,
		GzipLevel:         gcs.GzipLevel,
		Format:            gcs.Format,
		MessageType:       gcs.MessageType,
		ResponseCondition: gcs.ResponseCondition,
		TimestampFormat:   gcs.TimestampFormat,
		Placement:         gcs.Placement,
	}

	// Set new values if set by user.
	if c.NewName.Valid {
		input.NewName = c.NewName.Value
	}

	if c.Bucket.Valid {
		input.Bucket = c.Bucket.Value
	}

	if c.User.Valid {
		input.User = c.User.Value
	}

	if c.SecretKey.Valid {
		input.SecretKey = c.SecretKey.Value
	}

	if c.Path.Valid {
		input.Path = c.Path.Value
	}

	if c.Period.Valid {
		input.Period = c.Period.Value
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = c.FormatVersion.Value
	}

	if c.GzipLevel.Valid {
		input.GzipLevel = c.GzipLevel.Value
	}

	if c.Format.Valid {
		input.Format = c.Format.Value
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = c.ResponseCondition.Value
	}

	if c.TimestampFormat.Valid {
		input.TimestampFormat = c.TimestampFormat.Value
	}

	if c.MessageType.Valid {
		input.MessageType = c.MessageType.Value
	}

	if c.Placement.Valid {
		input.Placement = c.Placement.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	gcs, err := c.Globals.Client.UpdateGCS(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated GCS logging endpoint %s (service %s version %d)", gcs.Name, gcs.ServiceID, gcs.Version)
	return nil
}
