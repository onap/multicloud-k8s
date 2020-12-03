#!/bin/bash

# Precondition:
# QAT device installed, such as lspci -n | grep 37c8
# Enable grub with "intel_iommu=on iommu=pt"

ROOT=

MV=mv
RM=rm
ECHO=echo
SLEEP=sleep
INSTALL="/usr/bin/install -c"
MKDIR_P="mkdir -p"
libdir=/usr/local/lib
bindir=/usr/local/bin

am__append_1="drivers/crypto/qat/qat_dh895xcc/qat_dh895xcc.ko\
        drivers/crypto/qat/qat_dh895xccvf/qat_dh895xccvf.ko"

am__append_2="qat_895xcc.bin qat_895xcc_mmp.bin"
am__append_3="dh895xcc_dev0.conf dh895xcc_dev1.conf dh895xccvf_dev0.conf.vm"

# Kernel modules list
KO_MODULES_LIST="drivers/crypto/qat/qat_common/intel_qat.ko \
    drivers/crypto/qat/qat_c62x/qat_c62x.ko \
    drivers/crypto/qat/qat_c62xvf/qat_c62xvf.ko \
    drivers/crypto/qat/qat_d15xx/qat_d15xx.ko \
    drivers/crypto/qat/qat_d15xxvf/qat_d15xxvf.ko \
    drivers/crypto/qat/qat_c3xxx/qat_c3xxx.ko \
    drivers/crypto/qat/qat_c3xxxvf/qat_c3xxxvf.ko $am__append_1"

# Firmwares list
BIN_LIST="qat_c3xxx.bin qat_c3xxx_mmp.bin qat_c62x.bin \
    qat_c62x_mmp.bin qat_mmp.bin qat_d15xx.bin qat_d15xx_mmp.bin \
    $am__append_2"
CONFIG_LIST="c3xxx_dev0.conf \
    c3xxxvf_dev0.conf.vm \
    c6xx_dev0.conf \
    c6xx_dev1.conf \
    c6xx_dev2.conf \
    c6xxvf_dev0.conf.vm \
    d15xx_dev0.conf \
    d15xxpf_dev0.conf \
    d15xxvf_dev0.conf.vm \
    $am__append_3"

QAT_DH895XCC_NUM_VFS=32
QAT_DHC62X_NUM_VFS=16
QAT_DHD15XX_NUM_VFS=16
QAT_DHC3XXX_NUM_VFS=16

# Device information variables
INTEL_VENDORID="8086"
DH895_DEVICE_NUMBER="0435"
DH895_DEVICE_NUMBER_VM="0443"
C62X_DEVICE_NUMBER="37c8"
C62X_DEVICE_NUMBER_VM="37c9"
D15XX_DEVICE_NUMBER="6f54"
D15XX_DEVICE_NUMBER_VM="6f55"
C3XXX_DEVICE_NUMBER="19e2"
C3XXX_DEVICE_NUMBER_VM="19e3"
numC62xDevice=`lspci -vnd 8086: | egrep -c "37c8|37c9"`
numD15xxDevice=`lspci -vnd 8086: | egrep -c "6f54|6f55"`
numDh895xDevice=`lspci -vnd 8086: | egrep -c "0435|0443"`
numC3xxxDevice=`lspci -vnd 8086: | egrep -c "19e2|19e3"`
numDh895xDevicesP=`lspci -n | egrep -c "$INTEL_VENDORID:$DH895_DEVICE_NUMBER"`
numDh895xDevicesV=`lspci -n | egrep -c "$INTEL_VENDORID:$DH895_DEVICE_NUMBER_VM"`
numC62xDevicesP=`lspci -n | egrep -c "$INTEL_VENDORID:$C62X_DEVICE_NUMBER"`
numD15xxDevicesP=`lspci -n | egrep -c "$INTEL_VENDORID:$D15XX_DEVICE_NUMBER"`
numC3xxxDevicesP=`lspci -n | egrep -c "$INTEL_VENDORID:$C3XXX_DEVICE_NUMBER"`
MODPROBE_BLACKLIST_FILE="blacklist-qat-vfs.conf"

# load vfio-pci
$ECHO "Loading module vfio-pci"
modprobe vfio-pci

# qat-driver
$ECHO "Installing driver in `\uname -r`"
INSTALL_MOD_DIR=/lib/modules/`\uname -r`/updates/
for ko in $KO_MODULES_LIST; do
    base=${ko%/*};
    file=${ko##*/};
    mkdir -p $ROOT$INSTALL_MOD_DIR$base
    $INSTALL $file $ROOT$INSTALL_MOD_DIR$base
done

# qat-adf-ctl
if [ ! -d $ROOT$bindir ]; then
    $MKDIR_P $ROOT$bindir;
fi;
$INSTALL -D -m 750 adf_ctl $ROOT$bindir/adf_ctl;

# qat-service
if [ ! -d $ROOT/lib/firmware/qat_fw_backup ]; then
    $MKDIR_P $ROOT/lib/firmware/qat_fw_backup;
fi;

for bin in $BIN_LIST; do
    if [ -e $ROOT/lib/firmware/$bin ]; then
        mv $ROOT/lib/firmware/$bin $ROOT/lib/firmware/qat_fw_backup/$bin;
    fi;
    if [ -e $bin ]; then
        $INSTALL -D -m 750 $bin $ROOT/lib/firmware/$bin;
    fi;
done;

if [ ! -d $ROOT/etc/qat_conf_backup ]; then
    $MKDIR_P $ROOT/etc/qat_conf_backup;
fi;
$MV $ROOT/etc/dh895xcc*.conf $ROOT/etc/qat_conf_backup/ 2>/dev/null;
$MV $ROOT/etc/c6xx*.conf $ROOT/etc/qat_conf_backup/ 2>/dev/null;
$MV $ROOT/etc/d15xx*.conf $ROOT/etc/qat_conf_backup/ 2>/dev/null;
$MV $ROOT/etc/c3xxx*.conf $ROOT/etc/qat_conf_backup/ 2>/dev/null;

for ((dev=0; dev<$numDh895xDevicesP; dev++)); do
    $INSTALL -D -m 640 dh895xcc_dev0.conf $ROOT/etc/dh895xcc_dev$dev.conf;
    for ((vf_dev = 0; vf_dev<$QAT_DH895XCC_NUM_VFS; vf_dev++)); do
        vf_dev_num=$(($dev * $QAT_DH895XCC_NUM_VFS + $vf_dev));
        $INSTALL -D -m 640 dh895xccvf_dev0.conf.vm $ROOT/etc/dh895xccvf_dev$vf_dev_num.conf;
    done;
done;

for ((dev=0; dev<$numC62xDevicesP; dev++)); do
    $INSTALL -D -m 640 c6xx_dev$(($dev%3)).conf $ROOT/etc/c6xx_dev$dev.conf;
    for ((vf_dev = 0; vf_dev<$QAT_DHC62X_NUM_VFS; vf_dev++)); do
        vf_dev_num=$(($dev * $QAT_DHC62X_NUM_VFS + $vf_dev));
        $INSTALL -D -m 640 c6xxvf_dev0.conf.vm $ROOT/etc/c6xxvf_dev$vf_dev_num.conf;
    done;
done;

for ((dev=0; dev<$numD15xxDevicesP; dev++)); do
    $INSTALL -D -m 640 d15xx_dev$(($dev%3)).conf $ROOT/etc/d15xx_dev$dev.conf;
    for ((vf_dev = 0; vf_dev<$QAT_DHD15XX_NUM_VFS; vf_dev++)); do
        vf_dev_num=$(($dev * $QAT_DHD15XX_NUM_VFS + $vf_dev));
        $INSTALL -D -m 640 d15xxvf_dev0.conf.vm $ROOT/etc/d15xxvf_dev$vf_dev_num.conf;
    done;
done;

for ((dev=0; dev<$numC3xxxDevicesP; dev++)); do
    $INSTALL -D -m 640 c3xxx_dev0.conf $ROOT/etc/c3xxx_dev$dev.conf;
    for ((vf_dev = 0; vf_dev<$QAT_DHC3XXX_NUM_VFS; vf_dev++)); do
        vf_dev_num=$(($dev * $QAT_DHC3XXX_NUM_VFS + $vf_dev));
        $INSTALL -D -m 640 c3xxxvf_dev0.conf.vm $ROOT/etc/c3xxxvf_dev$vf_dev_num.conf;
    done;
done;

$ECHO "Creating startup and kill scripts";
if [ ! -d $ROOT/etc/modprobe.d ]; then
    $MKDIR_P $ROOT/etc/modprobe.d;
fi;
$INSTALL -D -m 750 qat_service $ROOT/etc/init.d/qat_service;
$INSTALL -D -m 750 qat_service_vfs $ROOT/etc/init.d/qat_service_vfs;
$INSTALL -D -m 750 qat $ROOT/etc/default/qat;
if [ -e $ROOT/etc/modprobe.d/$MODPROBE_BLACKLIST_FILE ] ; then
    $RM $ROOT/etc/modprobe.d/$MODPROBE_BLACKLIST_FILE;
fi;

if [ $numDh895xDevicesP != 0 ];then
    $ECHO "blacklist qat_dh895xccvf" >> $ROOT/etc/modprobe.d/$MODPROBE_BLACKLIST_FILE;
fi;
if [ $numC3xxxDevicesP != 0 ];then
    $ECHO "blacklist qat_c3xxxvf" >> $ROOT/etc/modprobe.d/$MODPROBE_BLACKLIST_FILE;
fi;
if [ $numC62xDevicesP != 0 ];then
    $ECHO "blacklist qat_c62xvf" >> $ROOT/etc/modprobe.d/$MODPROBE_BLACKLIST_FILE;
fi;
if [ $numD15xxDevicesP != 0 ];then
    $ECHO "blacklist qat_d15xxvf" >> $ROOT/etc/modprobe.d/$MODPROBE_BLACKLIST_FILE;
fi;

if [ ! -d $ROOT$libdir ]; then
    $MKDIR_P $ROOT$libdir;
fi;
if [ ! -d $ROOT/etc/ld.so.conf.d ]; then
    $MKDIR_P $ROOT/etc/ld.so.conf.d;
fi;
if [ ! -d $ROOT/lib/modules/`\uname -r`/kernel/drivers ]; then
    $MKDIR_P $ROOT/lib/modules/`\uname -r`/kernel/drivers;
fi;
if [ ! -d $ROOT/etc/udev/rules.d ]; then
    $MKDIR_P $ROOT/etc/udev/rules.d;
fi;

$ECHO "Copying libqat_s.so to $ROOT$libdir";
$INSTALL -D -m 755 libqat_s.so $ROOT$libdir/libqat_s.so;
$ECHO "Copying libusdm_drv_s.so to $ROOT$libdir";
$INSTALL -D -m 755 libusdm_drv_s.so $ROOT$libdir/libusdm_drv_s.so;
$ECHO $libdir > $ROOT/etc/ld.so.conf.d/qat.conf; ldconfig;

$ECHO "Copying usdm module to system drivers";
$INSTALL usdm_drv.ko "$ROOT/lib/modules/`\uname -r`/kernel/drivers";
$INSTALL qat_api.ko  "$ROOT/lib/modules/`\uname -r`/kernel/drivers";
$ECHO "Creating udev rules";
if [ ! -e $ROOT/etc/udev/rules.d/00-qat.rules ]; then
    echo 'KERNEL=="qat_adf_ctl" MODE="0660" GROUP="qat"' > $ROOT/etc/udev/rules.d/00-qat.rules;
    echo 'KERNEL=="qat_dev_processes" MODE="0660" GROUP="qat"' >> $ROOT/etc/udev/rules.d/00-qat.rules;
    echo 'KERNEL=="usdm_drv" MODE="0660" GROUP="qat"' >> $ROOT/etc/udev/rules.d/00-qat.rules;
    echo 'KERNEL=="uio*" MODE="0660" GROUP="qat"' >> $ROOT/etc/udev/rules.d/00-qat.rules;
    echo 'KERNEL=="hugepages" MODE="0660" GROUP="qat"' >> $ROOT/etc/udev/rules.d/00-qat.rules;
fi;
$ECHO "Creating module.dep file for QAT released kernel object";
$ECHO "This will take a few moments";
depmod -a;
if [ `lsmod | grep "usdm_drv" | wc -l` != "0" ]; then
    $ECHO "rmmod usdm_drv";
    rmmod usdm_drv;
fi;
if [ -e /sbin/chkconfig ] ; then
    chkconfig --add qat_service;
elif [ -e /usr/sbin/update-rc.d ]; then
    $ECHO "update-rc.d qat_service defaults";
    update-rc.d qat_service defaults;
fi;

$ECHO "Starting QAT service";
/etc/init.d/qat_service shutdown;
$SLEEP 3;
/etc/init.d/qat_service start;
/etc/init.d/qat_service_vfs start;

# load kernel vf module for QAT device plugin
numC62xDevicesV=`lspci -n | egrep -c "$INTEL_VENDORID:$C62X_DEVICE_NUMBER_VM"`
numD15xxDevicesV=`lspci -n | egrep -c "$INTEL_VENDORID:$D15XX_DEVICE_NUMBER_VM"`
numC3xxxDevicesV=`lspci -n | egrep -c "$INTEL_VENDORID:$C3XXX_DEVICE_NUMBER_VM"`
if [ $numC62xDevicesV != 0 ];then
    $ECHO "Loading qat_c62xvf";
    modprobe qat_c62xvf
fi
if [ $numC3xxxDevicesV != 0 ];then
    $ECHO "Loading qat_c3xxxvf";
    modprobe qat_c3xxxvf
fi
if [ $numD15xxDevicesV != 0 ];then
    $ECHO "Loading qat_d15xxvf";
    modprobe qat_d15xxvf
fi

