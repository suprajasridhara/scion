import argparse
import json
import time
from pathlib import Path
import shutil
import toml
import os

ADDR = "127.0.0.#count#"
port = 2000
count = 100
ID = "#ID#"
GEN_PATH="#GEN_PATH#"
ISD_ID="#ISD_ID#"
AS_ID_="#AS_ID_#"
SD_ADDR="#SD_ADDR#"
IP="#IP#"
PORT="#PORT#"
AS_ID="#AS_ID#"
QUIC_PORT="#QUIC_PORT#"
DB = "#DB#"

def gen_config(service, id, ia):
    global count
    global port
    ip = ADDR.replace("#count#", str(count))
    count = count + 1
    Path("./gen").mkdir(parents=True, exist_ok=True)
    with open("./config/"+service.lower()+".conf") as f:
        conf = f.read()

    conf = conf.replace(ID, id).replace(GEN_PATH, args.genpath).replace(ISD_ID, ia.split("-")[0]).replace(AS_ID,ia.split("-")[1]).replace(AS_ID_,ia.split("-")[1].replace(":","_")).replace(IP,ip).replace(PORT,str(port)).replace(DB, id+".db")
    port = port + 1
    conf = conf.replace(QUIC_PORT, str(port)).replace(SD_ADDR,get_sd_addr(ia))
    
    config_path = "./gen/"+id+".conf"
    with open(config_path, "w") as f:
        f.write(conf)
    return config_path

def get_sd_addr(ia):
    with open(args.genpath+"/ISD"+ia.split("-")[0]+"/AS"+ia.split("-")[1].replace(":","_")+"/endhost/sd.toml") as f:
        parsed_toml = toml.loads(f.read())
        return parsed_toml['sd']['address']

def clean():
    shutil.rmtree("./gen",ignore_errors=True)
    os.system("pkill -9 ms")
    os.system("pkill -9 pln")
    os.system("pkill -9 pgn")
    os.system("pkill -9 sig")

def start_service(service, config, instance):
    current = os.getcwd()
    os.chdir("./run")
    os.system("./r.sh "+service.lower()+" "+"."+config+" "+instance)
    os.chdir(current)

def build_all():
    current = os.getcwd()
    os.chdir("./bin")
    os.system("./build_all.sh")
    os.chdir(current)

clean()

parser = argparse.ArgumentParser(description='run tests')
parser.add_argument('--folder', type=str,
                    help='folder with test files and config', required=True)
parser.add_argument('--services', type=str,
                    help='services json files relative to folder', required=True)
parser.add_argument('--genpath', type=str,
                    help='path to the gen folder for the scion network', required=True)
args = parser.parse_args()
print("Build services")
build_all()
with open(args.folder+"/"+args.services) as f:
  services = json.load(f)

#for service in services:

service = 'PLN'
for instance in services['PLN']:
    ia=services[service][instance]['ia']
    config_path = gen_config(service, instance, ia)
    start_service(service, config_path, instance)

service = 'PGN'
for instance in services['PGN']:
    ia=services[service][instance]['ia']
    config_path = gen_config(service, instance, ia)
    start_service(service, config_path, instance)

service = 'MS'
for instance in services['MS']:
    ia=services[service][instance]['ia']
    config_path = gen_config(service, instance, ia)
    start_service(service, config_path, instance)

time.sleep(20)
service = 'SIG'
for instance in services['SIG']:
    ia=services[service][instance]['ia']
    config_path = gen_config(service, instance, ia)
    start_service(service, config_path, instance)
