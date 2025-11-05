//go:build !custom || secretstores || secretstores.gcloud

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/gcloud" // register plugin
