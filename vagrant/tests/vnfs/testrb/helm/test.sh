RED='\033[0;31m'
NC='\033[0m'

echo -e "${RED}Create Helm Chart Archive ---------------------------------------${NC}"
tar -cvf vault-consul-dev.tar vault-consul-dev/
gzip vault-consul-dev.tar

echo -e "${RED}Create Definition Metadata---------------------------------------${NC}"
curl -i -d @create_rbdefinition.json -X POST http://localhost:8081/v1/rb/definition

echo -e "${RED}Upload Definition data-------------------------------------------${NC}"
curl -i --data-binary @vault-consul-dev.tar.gz -X POST http://localhost:8081/v1/rb/definition/7eb09e38-4363-9942-1234-3beb2e95fd85/content

echo -e "${RED}LIST all Definitions---------------------------------------------${NC}"
curl -i -X GET http://localhost:8081/v1/rb/definition

echo -e "${RED}Create Profile Archive ------------------------------------------${NC}"
cd profile
tar -cvf profile.tar *
gzip profile.tar
cd ..
cp profile/profile.tar.gz .

echo -e "${RED}Create Profile Metadata------------------------------------------${NC}"
curl -i -d @create_rbprofile.json -X POST http://localhost:8081/v1/rb/profile

echo -e "${RED}Upload Profile data----------------------------------------------${NC}"
curl -i --data-binary @profile.tar.gz -X POST http://localhost:8081/v1/rb/profile/12345678-8888-4578-3344-987654398731/content

echo -e "${RED}LIST all Profiles------------------------------------------------${NC}"
curl -i -X GET http://localhost:8081/v1/rb/profile

echo -e "${RED}Instantiate Profile ---------------------------------------------${NC}"
curl -d @create_rbinstance.json http://localhost:8081/v1/vnf_instances/

echo -e "${RED}Delete Instantiation with following command----------------------${NC}"
echo "curl -X DELETE http://localhost:8081/v1/vnf_instances/krd/testnamespace1/<vnf_id>"