///* 
//* File:   longtooth.h
//* Author: robinshang
//*
//* Created on March 8, 2016, 15:37 PM
//*/

#ifndef LONGTOOTH_H
#define	LONGTOOTH_H
#include <stdlib.h>

// MSVC++ 9.0  _MSC_VER == 1500 (Visual Studio 2008)
#if (defined _MSC_VER) && (_MSC_VER <= 1500)
typedef short int int16_t;
typedef __int32 int32_t;
typedef unsigned __int32 uint32_t;
typedef __int64 int64_t;
typedef unsigned __int64 uint64_t;
typedef unsigned int size_t;
typedef int ssize_t;
typedef int bool;
typedef int in_addr_t;
typedef short int in_port_t;
#define true 1
#define false 0
#else
#include <stdint.h>
#include <stdbool.h>
#endif

//#define  LOG_OFF                -1
//#define  LOG_FATAL		0
//#define  LOG_ERROR		1
//#define  LOG_INFO		2
//#define  LOG_WARN		3
//#define  LOG_DEBUG		4
//#define  LOG_TRACE		5
//#define  LOG_ALL		100
//
//#define ARGUMENT_INT32  1
//#define ARGUMENT_INT64  2
//#define ARGUMENT_BYTE   3
//#define ARGUMENT_STRING 4
//#define ARGUMENT_BYTES  5

//#define EVENT_LONGTOOTH_STOPPED             0x20000
// 长牙内网功能已启动
#define EVENT_LONGTOOTH_STARTED             0x20001
// 长牙已完全启动
#define EVENT_LONGTOOTH_ACTIVATED           0x20002
// 长牙注册期间的事件码
#define EVENT_LONGTOOTH_REGISTER_SOCK_CONNECTED      0x20004
#define EVENT_LONGTOOTH_REGISTER_SOCK_ERROR   0x20005
#define EVENT_LONGTOOTH_SWITCH_SOCK_CONNECTED      0x20006
#define EVENT_LONGTOOTH_SWITCH_SOCK_ERROR   0x20007

// appid和key不匹配
#define EVENT_LONGTOOTH_INVALID             0x28001
// 请求或响应超时
#define EVENT_LONGTOOTH_TIMEOUT             0x28002
// 目标长牙ID无法访问
#define EVENT_LONGTOOTH_UNREACHABLE         0x28003
// 目标长牙ID不在线
#define EVENT_LONGTOOTH_OFFLINE             0x28004
#define EVENT_LONGTOOTH_BROADCAST           0x28005

// 目标长牙应用的服务不存在
#define EVENT_SERVICE_NOT_EXIST             0x40001
#define EVENT_SERVICE_INVALID               0x40002
#define EVENT_SERVICE_EXCEPTION             0x40003
#define EVENT_SERVICE_TIMEOUT               0x40004

#ifdef	__cplusplus
extern "C" {
#endif

// 长牙隧道仅使用数组参数通讯
// LT tunnel just use arguments to communicate
#define LT_ARGUMENTS		0
// 长牙隧道使用数组参数和数据流方式通讯
// LT tunnel use arguments and stream to communicate
#define LT_STREAM		1
// 长牙隧道仅使用数组参数和数据报文方式通讯
// LT tunnel use arguments and datagram to communicate
#define LT_DATAGRAM		2

    typedef char lt_tunnel[10];
	/**
	* \brief 设置长牙注册服务器
	* \param   host    长牙注册服务器地址
	* \param   port    长牙注册服务器端口
	*/
	 void lt_register_host_set(const char* host, int port);

	/**
	* @brief   获取本端LongToothID
	* @return  本端LongToothID字符串
	*/
	 const char* lt_id();

	/**
	* \brief   LongTooth附件处理函数,在LongTooth调用lt_event_handler和lt_service_response_handler回调函数时使用
	* \param   attachment  LongTooth附件
	* \param   ...         可变参数
	* \see lt_event_handler,lt_service_response_handler,lt_request,lt_respond
	*/
	 typedef void*(*lt_attachment_handler)(void* attachment, ...);

	/**
	* \brief   LongTooth事件处理函数
	* \param   event   长牙事件代码
	* \param   ltid_str    事件相关LongToothID
	* \param   service_str     事件相关LongTooth服务
	* \param   msg         事件相关信息
	* \param   attachment  事件相关附件,该附件在调用lt_request和lt_response时输入
	* \param   handler     附件处理函数
	* \see     EVENT_LONGTOOTH_STARTED,EVENT_LONGTOOTH_UNREACHABLE,EVENT_LONGTOOTH_OFFLINE,EVENT_SERVICE_NOT_EXIST
	* \see     lt_start
	*/
	 typedef void(*lt_event_handler)(int event, 
									const char* ltid_str, 
									const char* srv_str,
									const char* msg,
									size_t msglen,
									void* attachment, 
									lt_attachment_handler handler);

	/**
	* @brief   长牙启动
	* @param devid     长牙开发者ID
	* @param appid     长牙应用ID
	* @param appkey    长牙应用公钥
	* @param machineid 设备UUID,用来产生长牙实例－-长牙ID的唯一标示
	* @param handler   长牙事件处理函数
	* @return 0,启动成功;-1,启动失败
	* @see lt_event_handler
	*/
	int lt_start(int64_t devid, 
				int appid, 
				const char* appkey, 
				int64_t machineid, 
				lt_event_handler handler);

	/**
	* \brief   LongTooth服务请求处理函数
	* \param   ltt             服务请求方的长牙隧道
	* \param   ltid_str        服务请求方LongToothID
	* \param   service_str     请求LongTooth服务
	* \param   data_type       服务请求方的通讯方式
	* \param   args            服务请求方执行lt_request时输入的数组参数
	* \param   argslen         服务请求方的数组参数长度
	* \see     LT_ARGUMENTS,LT_STREAM,LT_DATAGRAM
	* \see     lt_service_add,lt_request
	*/
	 typedef void (*lt_service_request_handler)(const lt_tunnel ltt,
												const char* ltid_str,
												const char* service_str, 
												int data_type, 
												const char* args, 
												size_t argslen);

	/**
	* @brief   添加长牙服务，等待外部服务请求
	* @param service_str   长牙服务名，不能为空
	* @param req_handler   长牙服务请求处理函数，不能为空
	*/
	 void lt_service_add(const char* service_str, lt_service_request_handler req_handler);

	 int lt_broadcast(const char* keyword, 
					const char* msg, 
					int msg_len);

	/**
	* \brief   LongTooth服务响应处理函数
	* \param   ltt             服务提供者的长牙隧道
	* \param   ltid_str        服务提供者的LongToothID
	* \param   service_str     请求的LongTooth服务
	* \param   data_type       服务提供者的通讯方式
	* \param   args            服务提供者执行lt_repond的数组参数
	* \param   argslen         服务提供者的数组参数长度
	* \param   attachment      服务请求方执行lt_request时输入的附件
	* \param   ahandler        附件处理函数
	* \see     LT_ARGUMENTS,LT_STREAM,LT_DATAGRAM
	* \see     lt_request,lt_response
	*/   
	 typedef void (*lt_service_response_handler)(const lt_tunnel ltt, 
												const char* ltid_str,
												const char* service_str, 
												int data_type, 
												const char* args, 
												size_t argslen,
												void* attachment, 
												lt_attachment_handler ahandler);

	/**
	* @brief   长牙服务请求函数
	* @param ltt           空白长牙隧道,长牙服务请求执行后有效,不能为NULL
	* @param ltid_str      服务端长牙ID,不能为NULL
	* @param service_str   服务端服务名称,不能为NULL
	* @param lt_data_type  服务请求长牙隧道通讯方式
	* @param args          服务请求数组参数，长度不大于1024字节
	* @param argslen       服务请求数组参数长度
	* @param attachment    服务请求附件,在lt_service_response_handler和lt_event_handler中使用
	* @param a_handler     服务请求附件处理函数
	* @param resp_handler  服务响应处理函数,不能为NULL
	* @return 0,长牙服务请求执行;－1,未执行
	* @see     LT_ARGUMENTS,LT_STREAM,LT_DATAGRAM
	* @see     lt_service_response_handler
	*/
	int lt_request(lt_tunnel ltt,
					const char* ltid_str, 
					const char* service_str, 
					int lt_data_type, 
					const char* args,
					size_t argslen,
					void* attachment,
					lt_attachment_handler a_handler,
					lt_service_response_handler resp_handler);

	/**
	* @brief   服务响应函数
	* @param ltt           服务请求者的长牙隧道,不能为NULL
	* @param lt_data_type  响应服务请求的通讯方式
	* @param args          服务响应数组参数，长度不大于1024字节
	* @param argslen       服务响应数组参数长度
	* @param attachment    服务响应附件,在lt_event_handler中使用
	* @param a_handler     服务响应附件处理函数
	* @return 0,长牙服务响应执行;－1,未执行
	* @see     LT_ARGUMENTS,LT_STREAM,LT_DATAGRAM
	*/
	 int lt_respond(const lt_tunnel ltt, 
					int lt_data_type, 
					const char* args, 
					size_t argslen,
					void* attachment, 
					lt_attachment_handler a_handler);

	/**
	* @brief   从长牙隧道中读取数据
	* @param ltt   长牙隧道
	* @param buf   读取缓冲区,NULL表示结束读取数据
	* @param length    缓冲区长度,小于1表示结束读取数据
	* @return >0,读取的数据长度;=0,等待下次读取;=-1,读取结束
	*/
	 int ltt_receive(const lt_tunnel ltt, 
					char* buf, 
					int length);
    /**
     * @brief set receive timeout
     * @param timeout   the millseconds
     */
    void ltt_receive_timeout_set(int timeout);
    
	/**
	* @brief   向长牙隧道中发送数据
	* @param ltt   长牙隧道
	* @param buf   发送缓冲区,NULL表示结束发送数据
	* @param length    缓冲区长度,小于1表示结束发送数据
	* @return >0,发送的数据长度;=0,等待下次发送;=-1,发送结束
	*/
	 int ltt_send(const lt_tunnel ltt, 
				const char* buf, 
				int length);
    
    /**
     * @brief set send timeout
     * @param timeout   the millseconds
     */
    void ltt_send_timeout_set(int timeout);

	/**
	* @brief   设置最大线程数（默认线程数为16，最小为16）
	* @param max   线程数
	*/
    void lt_thread_max(int max);
    
	/**
	* @brief   设置内网是否可用，默认内网可用
	* @param enable   false:内网不可用; true:内网可用
	*/
    void lt_lan_mode_set(bool enable);

    size_t lt_tunnel_counts();
    //void lt_tunnels_limit(size_t limit);
    
    
#ifdef	__cplusplus
}
#endif

#endif	/* LONGTOOTH_H */

