# yp-diploma

Общая схема регистрации и обработки заказов:
1.После регистрации/авторизации в системе регистрируется номер заказа. Полученный номер заносится одновременно в БД и в пулл для обработки заказов.
2.Для обработки (получения баллов из системы начисления баллов) есть следующие компоненты:
jobPool : перечень заказов, которые ожидают обработки. jobPool безопасен для многопоточного доступа.
jobDispatcher: балансировщик нагрузки, каждый заданный интервал времени забирает из jobPool задачи, в количестве не превышающем ограничение и передает их в jobProcessor. Потом, получает результат от процессора и обрабатывает его заданной функцией (помещает в БД в нашем случае)
jobProcessor: управляет воркерами и распределяет между ними задания полученные от jobDispatcher. Собирает выполненные воркерами задания (результаты обращений в систему начисления баллов) и пачкой возвращает их в jobDispatcher
Worker: Nштук, выполняют в потоковом режиме переданную функцию (обращения в систему начисления баллов) с полученным от Processor'а job, и возвращает результат или возникшую ошибку при выполненнии job.
3. В случае получения от системы начисления баллов ответа 200 со статсусом REGISTERED/PROCESSING или ответ 204, заказ остается в пулле обработки для следующей итерации  
4. В случае получения от системы начисления баллов ответа 200 со статсусом INVALID/PROCESSED, заказу устанавливается начисленное число баллов и он сохраняется в БД.
5. При запуске сервиса проводится проверка зарегистрированных, но не направленных в систему начисления баллов заказов, и в случае,если такие есть они загружаются в пулл обработки и далее они обрабатываются по п.2-4

Сделано:
1. Структура таблиц бд. (пока в repository.go "DDL") 
2. Регистрация пользователя (/api/user/register), выдается кука с сессионным ключом
3. Авторизация пользователя (/api/user/login), выдается кука с сессионным ключом
4. Проверка сессионного ключа в мидлваре.
5. Поддержка gzip в мидлваре.
6. Регистрация заказа (/api/user/orders), номер заказ проверяется по алг.Луна, заносится в БД.
7. Получение информации о заказах покупателя (/api/user/orders). 
8. Получение информации о балансе покупателя (/api/user/balance).
9. Запрос на списание начисленных баллов (/api/user/balance/withdraw).
10.Получение информации о проведенных списаниях (/api/user/withdrawals)
11.(План):Вместо обращения к ситеме начисления баллов пока заглушка.
