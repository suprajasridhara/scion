# if ["$1" = "sig"]; then
#     sudo setcap cap_net_admin+ep $1
# fi
chmod +x ../bin/$1
nohup ../bin/$1 -config $2 > logs/$3.log &

echo $3 + "started"