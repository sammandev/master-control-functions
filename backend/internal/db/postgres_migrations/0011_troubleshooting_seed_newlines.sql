UPDATE troubleshooting_steps
SET script_snippet = E'[SSH]\nPROMPT=root@OpenWrt:~#'
WHERE guide_id = 1
  AND position = 2
  AND script_snippet = '[SSH]\nPROMPT=root@OpenWrt:~#';

UPDATE troubleshooting_steps
SET script_snippet = E'[DUT]\nIP=192.168.1.1\n; Referenced in script as *V:DUT-IP#'
WHERE guide_id = 3
  AND position = 1
  AND script_snippet = '[DUT]\nIP=192.168.1.1\n; Referenced in script as *V:DUT-IP#';

UPDATE troubleshooting_steps
SET script_snippet = E'[TX_5G_HE20]\nPOWER=-1.5~1.5\nEVM=-99~-35'
WHERE guide_id = 4
  AND position = 2
  AND script_snippet = '[TX_5G_HE20]\nPOWER=-1.5~1.5\nEVM=-99~-35';

UPDATE troubleshooting_steps
SET script_snippet = E'[RF_PATH]\nPORT_ANT1=RF1A\nPATH_LOSS_2G=4.2'
WHERE guide_id = 5
  AND position = 2
  AND script_snippet = '[RF_PATH]\nPORT_ANT1=RF1A\nPATH_LOSS_2G=4.2';
