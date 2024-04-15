FROM centos:7.6.1810
LABEL maintainer="https://www.idss-cn.com"

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && yum -y install kde-l10n-Chinese \
    && yum -y install glibc-common \
    && localedef -i zh_CN -f UTF-8 zh_CN.UTF-8 \
    && echo 'export LANG="zh_CN.UTF-8"' > /etc/locale.conf \
	&& echo 'export LC_ALL="zh_CN.UTF-8"' > /etc/locale.conf \
    && source /etc/locale.conf \
    && yum clean all 
ENV LANG=zh_CN.UTF-8 \
    LC_ALL=zh_CN.UTF-8

#安装oracle客户端
WORKDIR /tmp
COPY ./oracle-instantclient11.2-basic-11.2.0.4.0-1.x86_64.rpm /tmp/oracle-instantclient11.2-basic-11.2.0.4.0-1.x86_64.rpm
COPY ./oracle-instantclient11.2-sqlplus-11.2.0.4.0-1.x86_64.rpm /tmp/oracle-instantclient11.2-sqlplus-11.2.0.4.0-1.x86_64.rpm

RUN yum -y install oracle-instantclient11.2*

ENV ORACLE_BASE=/usr/lib/oracle/11.2 \
    ORACLE_HOME=/usr/lib/oracle/11.2/client64 \
    LD_LIBRARY_PATH=/usr/lib/oracle/11.2/client64/lib
#客户端字符集需参考服务端字符集，非必选
#ENV NLS_LANG="SIMPLIFIED CHINESE_CHINA".AL32UTF8

RUN ln -s /usr/lib/oracle/11.2/client64/bin/sqlplus /usr/bin \
    && mkdir -p /usr/lib/oracle/11.2/client64/network/admin

#启动采集器
WORKDIR /app/conf
COPY ./datacollector /app/datacollector
COPY ./datacollector.yaml /app/conf/datacollector.yaml
RUN chmod +x -R /app/datacollector
# CMD [ "/app/datacollector","-conf","/app/conf/datacollector.yaml"]
