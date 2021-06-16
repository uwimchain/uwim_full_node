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

Мнемофраза нужна для подписи транзакций, которые будут отправляться с вашей ноды. Ммнемофразу можно сгенерировать на сайте <https://wallet.uwim.io/><br>
**NODE_MNEMONIC=""**

Порт для API вашей ноды, не забудьте открыть его на сервере, на который ставите ноду<br>
**API_PORT=""**

Ip сервера, на который вы ставите ноду<br>
**NODE_IP=""**

После того, как вы установили Golang, установили и скомпилировали сборку ноды, а так же заполнили файл **config.env**, вы можете запустить ноду, тогда, если вы всё сделали правильно - нода свяжется с первым пиром-валидатором сети (это займёт некоторое время). После завершения подкачки, нода сможет полноценно функционировать.

Следите за обновлениями ноды

Для обновления ноды вам не обходимо зайдите в директорию, в которую вы установили ноду ранее и прописать команду:<br>
**git clone https://github.com/uwimchain/uwim_full_node.git**<br>

После этого пропишите команду:<br>
**go build**<br>

Далее снова запустите ноду, она подгрузит недостающие блоки и снова будет работать в стандартном режиме.
