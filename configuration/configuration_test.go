package configuration_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/blueprintue/discord-bot/configuration"
	"github.com/stretchr/testify/require"
)

func TestReadConfiguration(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"config.json": {
			Data: []byte(`{"discord": {"name": "foo","token": "bar"},"log": {"filename": "oof"}}`),
		},
	}

	expected := &configuration.Configuration{
		Discord: configuration.Discord{
			Name:  "foo",
			Token: "bar",
		},
		Log: configuration.Log{
			Filename:            "oof",
			Level:               "",
			NumberFilesRotation: 0,
		},
	}

	actualConfiguration, actualErr := configuration.ReadConfiguration(fsys, "config.json")

	require.NoError(t, actualErr)
	require.Equal(t, expected, actualConfiguration)
}

// because using t.SetEnv, no `t.Parallel()` allowed here.
func TestReadConfigurationWithEnvValues(t *testing.T) {
	fsys := fstest.MapFS{
		"config.json": {
			Data: []byte(`{"discord": {"name": "foo","token": "bar"},"log": {"filename": "oof", "level": "invalid", "number_files_rotation": 9}}`),
		},
	}

	t.Setenv("DBOT_DISCORD_NAME", "env_foo")
	t.Setenv("DBOT_DISCORD_TOKEN", "env_bar")
	t.Setenv("DBOT_LOG_FILENAME", "env_oof")
	t.Setenv("DBOT_LOG_LEVEL", "panic")
	t.Setenv("DBOT_LOG_NUMBER_FILES_ROTATION", "15")

	expected := &configuration.Configuration{
		Discord: configuration.Discord{
			Name:  "env_foo",
			Token: "env_bar",
		},
		Log: configuration.Log{
			Filename:            "env_oof",
			Level:               "panic",
			NumberFilesRotation: 15,
		},
	}

	actualConfiguration, actualErr := configuration.ReadConfiguration(fsys, "config.json")

	require.NoError(t, actualErr)
	require.Equal(t, expected, actualConfiguration)
}

//nolint:funlen
func TestReadConfigurationErrors(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"my_directory": {
			Mode: fs.ModeDir,
		},
		"config_invalid.json": {
			Data: []byte("foobar"),
		},
		"config_missing_discord_name.json": {
			Data: []byte(`{"discord": {"name": "","token": ""},"log": {"filename": ""}}`),
		},
		"config_missing_discord_token.json": {
			Data: []byte(`{"discord": {"name": "foo","token": ""},"log": {"filename": ""}}`),
		},
		"config_missing_log_filename.json": {
			Data: []byte(`{"discord": {"name": "foo","token": "bar"},"log": {"filename": ""}}`),
		},
		"config_invalid_log_level.json": {
			Data: []byte(`{"discord": {"name": "foo","token": "bar"},"log": {"filename": "rab", "level": "none"}}`),
		},
	}

	type args struct {
		filename string
	}

	type want struct {
		errorMessage string
	}

	testCases := map[string]struct {
		args args
		want want
	}{
		"should return error when no file provided": {
			args: args{
				filename: "",
			},
			want: want{
				errorMessage: "open : file does not exist",
			},
		},
		"should return error when file provided is invalid": {
			args: args{
				filename: "foobar",
			},
			want: want{
				errorMessage: "open foobar: file does not exist",
			},
		},
		"should return error when file provided is directory": {
			args: args{
				filename: "my_directory",
			},
			want: want{
				errorMessage: "read my_directory: invalid argument",
			},
		},
		"should return error when file provided is invalid json": {
			args: args{
				filename: "config_invalid.json",
			},
			want: want{
				errorMessage: "invalid character 'o' in literal false (expecting 'a')",
			},
		},
		"should return error when config missing Discord.Name": {
			args: args{
				filename: "config_missing_discord_name.json",
			},
			want: want{
				errorMessage: "invalid json value: discord.name is empty",
			},
		},
		"should return error when config missing Discord.Token": {
			args: args{
				filename: "config_missing_discord_token.json",
			},
			want: want{
				errorMessage: "invalid json value: discord.token is empty",
			},
		},
		"should return error when config missing Log.Filename": {
			args: args{
				filename: "config_missing_log_filename.json",
			},
			want: want{
				errorMessage: "invalid json value: log.filename is empty",
			},
		},
		"should return error when config has invalid Log.Level": {
			args: args{
				filename: "config_invalid_log_level.json",
			},
			want: want{
				errorMessage: "invalid json value: log.level is invalid",
			},
		},
	}

	for testCaseName, testCase := range testCases {
		testCaseName, testCase := testCaseName, testCase

		t.Run(testCaseName, func(tt *testing.T) {
			tt.Parallel()

			actualConfiguration, actualErr := configuration.ReadConfiguration(fsys, testCase.args.filename)

			require.ErrorContains(tt, actualErr, testCase.want.errorMessage)
			require.Nil(tt, actualConfiguration)
		})
	}
}
