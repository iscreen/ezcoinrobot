# EZCoinRobot

This is a microservice using gRPC protocal to operate supervisor services.

# Requisites
  * GO 1.16 above
  * protobuf 3.15 above
  * gRPC 


## Generate gRPC code
 
```bash
protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
protos/ezcoinrobot.proto
```

# Build Project

## Build client and server
  `make all`

# Systemd service

  - Create service name `/etc/systemd/system/ezcoinrobot.service` include below content.
    ``` 
    [Unit]
    Description=EZCoinService gRPC service
    After=network.target

    StartLimitIntervalSec=600
    StartLimitBurst=10

    [Service]
    Type=simple
    WorkingDirectory=/home/ubuntu/ezcoinrobot
    ExecStart=/home/ubuntu/ezcoinrobot/ezcoinrobot-server
    ExecStop=/bin/kill -TERM $MAINPID
    ExecReload=/bin/kill -TSTP $MAINPID

    # User=xxx
    # Group=xxx
    UMask=0002

    #LimitNOFILE=16384 # increase open files limit (see OS Tuning guide)

    Restart=on-failure
    RestartSec=10s
    SyslogIdentifier=ezcoin

    # Configure server using env vars
    [Install]
    WantedBy=multi-user.target
    ```
  - Start systemd service
    ```
    sudo systemctl enable ezcoinrobot.service
    sudo systemctl start ezcoinrobot.service
    ```