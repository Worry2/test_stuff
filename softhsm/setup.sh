mkdir -p /home/pasi/softhsm/tokens
export SOFTHSM2_CONF=$(pwd)/softhsm2.conf
softhsm2-util --init-token --slot 0 --label token
