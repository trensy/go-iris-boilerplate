FROM loads/alpine:3.8

LABEL maintainer="hpu423@126.com"

###############################################################################
#                                INSTALLATION
###############################################################################

# 设置固定的项目路径
ENV WORKDIR /var/www/http

RUN mkdir -p $WORKDIR
# 添加应用可执行文件，并设置执行权限
ADD bin/main $WORKDIR
RUN chmod +x  $WORKDIR

# 添加静态文件、配置文件、模板文件
ADD resource $WORKDIR/resource

###############################################################################
#                                   START
###############################################################################
WORKDIR $WORKDIR
CMD ./main -c $WORKDIR/resource/config/app.toml