# uwim chain golang node


Установка ноды:<br> 
1. Ubuntu 20.04 (рекомендуемо) или выше.
2. Golang 1.13+<br>
    **sudo apt update**<br>
    **sudo apt install golang**<br>

3. Скачать последнюю версию сборки ноды с **GitHub**, для этого вам нужно:

    3.1. Создать директорию, в которой будут храниться файлы вашей ноды
    
    3.2. Если у вас на сервере не установлен **Git**, то вам нужно его установить прописав команду:<br>
    	**sudo apt update**<br>
    	**sudo apt install git**<br>
    
    3.3. Проверить версию **Git** вы можете вписать команду:<br>
    	**git --version**<br>
    
    3.4. После установки **Git** перейдите в ранее созданую директорию и скачайте сборку ноды командой:<br>
    	**git clone https://github.com/uwimchain/uwim_full_node.git**<br>
    
После того, как вы установили Golang и скачали сборку, вам необходимо в терминале зайти в директорию, в которую вы установили сборку и прописать команду:<br> 
   **go build**

Эта команда установит дополнительные пакеты и библиотеки для работы ноды, а так же скомпилирует сборку в исполняемый файл **node**.

После проделанных действий вам нужно заполнить файл **config.env**:

Мнемофраза нужна для подписи транзакций, которые будут отправляться с вашей ноды. Мнемофразу можно сгенерировать на сайте <https://wallet.uwim.io/><br>
**NODE_MNEMONIC=""**

Порт для API вашей ноды, не забудьте открыть его на сервере, на который ставите ноду<br>
**API_PORT=""**

Ip сервера, на который вы ставите ноду<br>
**NODE_IP=""**

После того, как вы установили Golang, установили и скомпилировали сборку ноды, а так же заполнили файл **config.env**, вы можете запустить ноду, тогда, если вы всё сделали правильно - нода свяжется с первым пиром-валидатором сети (это займёт некоторое время). После завершения подкачки, нода сможет полноценно функционировать.

Следите за обновлениями ноды

Для обновления ноды вам необходимо зайти в директорию, в которую вы установили ноду ранее и прописать команду:<br>
**git clone https://github.com/uwimchain/uwim_full_node.git**<br>

После этого пропишите команду:<br>
**go build**<br>

Далее снова запустите ноду, она подгрузит недостающие блоки и снова будет работать в стандартном режиме.

<br><br>
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
