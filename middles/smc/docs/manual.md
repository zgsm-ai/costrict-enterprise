# smc Command Manual

## Introduction

smc is a command line tool provided by the smc platform for managing Data Store.

## Usage

### 1. Platform Operations

#### 1.1. Login

```sh
# Login to AI platform
smc login --addr 10.72.1.21 --username user100000 --password xxxx
```

#### 1.2. Logout

```shell
# Logout and remove local login cookie
smc logout
```

#### 1.3. Platform Permissions

```shell
smc grant --user zbc91868 --role updater
```

### 2. Dataset Operations

#### 2.1. Create

Create empty dataset:

```shell
smc dataset create --dataset sfdl/xdr.endpoint_behavior --type raw --schema behavior.json
smc dataset create --dataset sfdl/xdr.endpoint_behavior --type feature --schema labels.json
```

Upload data:

```shell
# Push label data file (behavior.csv), metadata (labels.json) and all files under ./images to AI platform 
# to build unstructured dataset xsec/obj_detect
# xsec/obj_detect refers to obj_detect dataset in xsec namespace
smc push --dataset xsec/obj_detect --files ./images --label behavior.csv --format csv
# Push data file (behavior.csv) and metadata (behavior.json) to AI platform 
# to build structured dataset sfdl/xdr.endpoint_behavior
smc push --dataset sfdl/xdr.endpoint_behavior --data behavior.csv --format csv
```

#### 2.2. Delete

```shell
# Delete dataset sfdl/xdr.endpoint_behavior
smc delete --dataset sfdl/xdr.endpoint_behavior
```

#### 2.3. Clone Dataset

```sh
# Copy to same namespace
smc clone --dataset sfdl/xdr.endpoint_behavior --target test
# Copy to specified namespace with new dataset name
smc clone --dataset sfdl/images --target zbctest/images
```

#### 2.4. Set Permissions

```shell
smc grant --dataset sfdl/xdr.endpoint_behavior --def-role browser
# Set user zbc as reader role for dataset sfdl/xdr.endpoint_behavior, with read access
smc grant --dataset sfdl/xdr.endpoint_behavior --user zbc reader
#smc grant --dataset sfdl/xdr.endpoint_behavior --group xdr --role updater
```

#### 2.5. Set Dataset Metadata

```shell
smc config --dataset sfdl/xdr.endpoint_behavior --key table.desc --value "Endpoint behavior data"
smc config --dataset sfdl/xdr.endpoint_behavior --key table.tag --value "XDR DataStore"
```

#### 2.6. View Dataset Info

```shell
# Get metadata for dataset sfdl/xdr.endpoint_behavior
smc view --dataset sfdl/xdr.endpoint_behavior
```

#### 2.7. Pull Dataset Data

```shell
# Pull dataset sfdl/xdr.endpoint_behavior to local directory ./behavior
smc pull --dataset sfdl/xdr.endpoint_behavior --outdir ./behavior
```

### 3. Namespace Operations [Admin]

#### 3.1. Create Namespace

```shell
smc dataspace create --dataspace sfdl --diskCapacity 20TB --datasetCapacity 10
```

#### 3.2. Delete Namespace

```shell
smc delete --dataspace sfdl
```

#### 3.3. Set Namespace Permissions [owner]

```shell
smc grant --dataspace sfdl --def-role browser
smc grant --dataspace sfdl --user zbc --role reader
# Group not implemented yet
#smc grant --dataspace sfdl --group xdr --role updater
```

#### 3.4. Set Namespace Properties

```shell
smc config --dataspace sfdl --key diskCapacity --value 1000TB
smc config --dataspace sfdl --key datasetCapacity --value 10
```

#### 3.5. View Namespace Info [owner]

```shell
smc view --dataspace sfdl
```

### 4. User Operations [Admin]

#### 4.1. Create User

```shell
# Create a user
smc user create --user zbc91868
# Create user with default platform role
smc user create --user zbc91868 --role updater
# Create user with namespace (1000TB disk quota, 10 datasets)
smc user create --user zbc91868 --diskCapacity 1000TB --datasetCapacity 10
```

#### 4.2. Delete User

```shell
smc delete --user zbc91868
```

#### 4.3. Set User Properties

```shell
smc config --user zbc91868 --key comment --value "Zheng Baichun"
#smc config --user zbc91868 --key group --value xdr
```

#### 4.4. View User Metadata

```shell
smc view --user zbc91868
# Other formats not supported yet
#smc view --user zbc91868 --output yaml
```

### 5. User Group Operations [Admin] (Not Supported Yet)

#### 5.1. Create Group

```shell
smc create --group XDR --desc "XDR Product Line"
```

#### 5.2. Delete Group

```shell
smc delete --group XDR
```

#### 5.3. Set Group Properties

```shell
smc config --group XDR --key desc --value "XDR Product Line"
```

#### 5.4. Add/Remove Group Members

```shell
smc add --group XDR --users zbc91868 tangming73136
smc rm --group XDR --users zbc91868
```

#### 5.4. View Group Metadata

```shell
smc view --group XDR
```

## Metadata

### 1. dataset

```json
{
    "dataset": {
        "dataspace": "sfdl",
        "dataset": "xdr.endpoint_behavior",
        "owner": "zbc91868",
        "type": "raw",
        "comment": "Endpoint behavior"
    },
    "status": {
        "diskUsed": "2222MB",
        "createTime": "2022-06-02 17:00:01", 
        "updateTime": "2022-06-03 23:59:59"
    }
}
```

### 2. dataspace

### 3. user


### Mount Datastore Namespace and Dataset

#### Mount

```sh
smc dataset mount beta/test
```

#### Switch Mount Engine

```sh
smc env -k smc_MOUNT_APP -v goofys
smc env -k smc_MOUNT_APP -v s3fs
smc env -k smc_MOUNT_APP -v rclone
```

Built-in support for three engines: goofys, s3fs, rclone

#### Add New Mount Engine

To extend mount engines, for example adding a new engine called "nb":

##### 1. Configure nb Engine Command

```sh
vim /etc/.smc/mount.nb.conf
```

Example mount.nb.conf:

```json
{
    "enabled": true,
    "command": "rclone mount --daemon -v --debug-fuse --config /etc/rclone/rclone.conf --uid 1000 --gid 1000 smc:smc/$name $mntdir",
    "debugCommand": "rclone mount --daemon --config /etc/rclone/rclone.conf --uid 1000 --gid 1000 smc:smc/$name $mntdir",
    "setuid": false
}
```

- enabled: whether this engine is enabled
- command: normal mode mount command
- debugCommand: debug mode mount command
- setuid: whether setuid elevation is needed

Command supports 4 predefined parameters:
- $s3url: S3 server URL (HTTP URL)  
- $name: dataset or namespace path
- $mntdir: target mount directory
- $home: current user's home directory

| Parameter | Meaning | Example |
|-----------|---------|---------|
| $s3url | S3 server URL | `http://10.72.1.225:12001` |
| $name | Dataset path | datalake-pulsar/http_log |
| $mntdir | Mount target | /root/datasets/datalake-pulsar/http_log |
| $home | Home directory | /home/jovyan |

##### 2. Enable nb Engine

```sh
smc env -k smc_MOUNT_APP -v nb
```

##### 3. Mount Dataset with nb Engine

```sh
smc dataset mount beta/ds04
```

##### 4. Check Mount Result

```sh
ls -l /root/datasets/beta/ds04
