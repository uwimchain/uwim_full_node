# uwim chain golang node

Node installation:<br> 
1. Ubuntu 20.04 (recommended) or higher.
2. Golang 1.13+<br>
    **sudo apt update**<br>
    **sudo apt install golang**<br>

3. To download the latest node assembly from **GitHub**, you need:

    3.1. Create a directory where your node files will be stored
    
    3.2. If you do not have **Git**, installed on your server, then you need to install it by writing the command:<br>
    	**sudo apt update**<br>
    	**sudo apt install git**<br>
    
    3.3. You can enter the command: **git --version** to check your **Git** version<br>
    
    3.4. After installing **Git**, go to the previously created directory and download the assembly of the node with the command:<br>
    	**git clone https://github.com/uwimchain/uwim_full_node.git**<br>
    
After you have installed Golang and downloaded the assembly, you need to go to the directory in the terminal where you installed the assembly and write the command:<br> 
   **go build**

This command will install additional packages and libraries for the node and compile the assembly into an executable file **node**.

After these steps, you need to fill in the **config.env** file:

Mnemophrase is needed to sign transactions that will be sent from your node. The mnemonic phrase can be generated on the website <https://wallet.uwim.io/><br>
**NODE_MNEMONIC=""**

Port for the API of your node, do not forget to open it on the server where you put the node<br>
**API_PORT=""**

Ip of the server where you put the node<br>
**NODE_IP=""**

You can start the node after you have installed Golang, installed and compiled the assembly of the node, and filled in the **config.env** file. If you did everything correctly, the node will contact the first peer-validator of the network (this will take some time). After completing the swap, the node will be able to fully function. 

Follow the node updates 

To update the node, you need to go to the directory where you installed the node earlier and write the command:<br>
**git clone https://github.com/uwimchain/uwim_full_node.git**<br>

After that, write the command:<br>
**go build**<br>

Next, run the node again, it will load the missing blocks and will work in standard mode again.
