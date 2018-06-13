extern void LtStartCallback(int event,
                              const char* ltid_str,
                              const char* srv_str,
                              const char* msg,
                              size_t msglen,
                              void* attachment,
                              lt_attachment_handler handler);

extern void LtRequestCallback(const char* ltt,
                                    const char *ltid_str,
                                    const char* service_str,
                                    int data_type,
                                    const char *args,
                                    int argslen,
                                    void* attachment,
                                    lt_attachment_handler ahandler);

int cb_lt_start(int64_t devid,
            int appid,
            const char* appkey,
            int64_t machineid);