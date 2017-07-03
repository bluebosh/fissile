#!/bin/bash

set -e

if [[ "$1" == "--help" ]]; then
cat <<EOL
Usage: run.sh
EOL
exit 0
fi

# Make BOSH installed binaries available
export PATH=/var/vcap/bosh/bin:$PATH

# Load RVM
source /usr/local/rvm/scripts/rvm

# Taken from https://github.com/cloudfoundry/bosh-linux-stemcell-builder/blob/95aa0de0fe734547b2dd9241685c31c5f6d61a83/stemcell_builder/lib/prelude_apply.bash
# To be used by scripts that are run or sourced by this file.
function get_os_type {
  centos_file=$chroot/etc/centos-release
  rhel_file=$chroot/etc/redhat-release
  ubuntu_file=$chroot/etc/lsb-release
  photonos_file=$chroot/etc/photon-release
  opensuse_file=$chroot/etc/SuSE-release

  os_type=''
  if [ -f $photonos_file ]
  then
    os_type='photonos'
  elif [ -f $ubuntu_file ]
  then
    os_type='ubuntu'
  elif [ -f $centos_file ]
  then
    os_type='centos'
  elif [ -f $rhel_file ]
  then
    os_type='rhel'
  elif [ -f $opensuse_file ]
  then
    os_type='opensuse'
  fi

  echo $os_type
}
export -f get_os_type

# Unmark the role. We may have this file from a previous run of the
# role, i.e. this may be a restart. Ensure that we are not seen as
# ready yet.
rm -f /var/vcap/monit/ready /var/vcap/monit/ready.lock

# When the container gets restarted, processes may end up with different pids
find /run -name "*.pid" -delete
if [ -d /var/vcap/sys/run ]; then
    find /var/vcap/sys/run -name "*.pid" -delete
fi

# Write a couple of identification files for the stemcell
mkdir -p /var/vcap/instance
echo {{ .role.Name }} > /var/vcap/instance/name
# TODO: we need to discover the index of the instance for HA
# things to work
echo 0 > /var/vcap/instance/id

# Note, any changes to this list of variables have to be replicated in
# --> model/mustache.go, func builtins
export IP_ADDRESS=$(/bin/hostname -i | awk '{print $1}')
export DNS_RECORD_NAME=$(/bin/hostname)

# Run custom environment scripts (that are sourced)
{{ range $script := .role.EnvironScripts }}
    source {{ if not (is_abs $script) }}/opt/hcf/startup/{{ end }}{{ $script }}
{{ end }}
# Run custom role scripts
{{ range $script := .role.Scripts}}
    bash {{ if not (is_abs $script) }}/opt/hcf/startup/{{ end }}{{ $script }}
{{ end }}

configgin \
	--jobs /opt/hcf/job_config.json \
	--env2conf /opt/hcf/env2conf.yml

if [ -e /etc/monitrc ]
then
  chmod 0600 /etc/monitrc
fi

# Create run dir
mkdir -p /var/vcap/sys/run
chown root:vcap /var/vcap/sys/run
chmod 775 /var/vcap/sys/run

# Fix permissions
chmod 640 /var/log/messages
chmod 1730 /var/spool/cron/crontabs/

{{ if eq .role.Type "bosh-task" }}
    # Start rsyslog and cron
    /usr/sbin/rsyslogd
    cron
{{ else }}
    # rsyslog and cron are started via monit
{{ end }}

# Run custom post config role scripts
# Run any custom scripts other than pre-start
{{ range $script := .role.PostConfigScripts}}
{{ if not (is_pre_start $script) }}
    echo bash {{ if not (is_abs $script) }}/opt/hcf/startup/{{ end }}{{ $script }}
    bash {{ if not (is_abs $script) }}/opt/hcf/startup/{{ end }}{{ $script }}
{{ end }}
{{ end }}

# Run all the scripts called pre-start, but ensure consul_agent/bin/pre-start is run before others.
# None of the other pre-start scripts appear to have any dependencies on one another.
function sorted-pre-start-paths()
{
    declare -a fnames
    idx=0
    if [ -x /var/vcap/jobs/consul_agent/bin/pre-start ] ; then
	fnames[$idx]=/var/vcap/jobs/consul_agent/bin/pre-start
	idx=$((idx + 1))
    fi
    for fname in $(find /var/vcap/jobs/*/bin -name pre-start | grep -v '/consul_agent/bin/pre-start$') ; do
	fnames[$idx]=$fname
	idx=$((idx + 1))
    done
    echo ${fnames[*]}
}

for fname in $(sorted-pre-start-paths) ; do
    echo bash $fname
    bash $fname
done

# Run
{{ if eq .role.Type "bosh-task" }}
    {{ range $job := .role.Jobs}}
        /var/vcap/jobs/{{ $job.Name }}/bin/run
    {{ end }}
{{ else }}

  killer() {
    # Wait for all monit services to be stopped
    echo "Received SIGTERM. Will run 'monit stop all'."

    total_services=$(monit summary | grep -c "^Process")

    monit stop all

    echo "Ran 'monit stop all'."

    while [ $total_services != $(monit summary | grep "^Process" | grep -c "Not monitored") ] ; do
       sleep 1
    done

    echo "All monit processes have been stopped."
    monit summary
    monit quit
  }

  trap killer SIGTERM

  if [[ "${LOG_LEVEL}" == "debug"* || -n "${LOG_DEBUG}" ]]; then
    # monit -v without the -I would fork a child, but then we can't wait on it,
    # so it's not very useful.
    monit -vI &
  else
    monit -I &
  fi
  child=$!
  wait "$child"
{{ end }}
