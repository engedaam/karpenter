MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="BOUNDARY"

--BOUNDARY
Content-Type: text/x-shellscript; charset="us-ascii"

#!/bin/bash
mkdir -p /etc/systemd/logind.conf.d
cat << EOF > /etc/systemd/logind.conf.d/50-max-delay.conf
[Login]
InhibitDelayMaxSec=360
EOF

systemctl restart systemd-logind

sed -i '/"apiVersion*/a \ \ "shutdownGracePeriod": "3m",' /etc/kubernetes/kubelet/kubelet-config.json
sed -i '/"shutdownGracePeriod*/a \ \ "shutdownGracePeriodCriticalPods": "2m",' /etc/kubernetes/kubelet/kubelet-config.json

--BOUNDARY
Content-Type: text/x-shellscript; charset="us-ascii"

#!/bin/bash
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1


echo $(jq '.containerLogMaxFiles=3|.containerLogMaxSize="100Mi"' /etc/kubernetes/kubelet/kubelet-config.json) > /etc/kubernetes/kubelet/kubelet-config.json

--BOUNDARY--