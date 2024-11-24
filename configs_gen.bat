set WORKSPACE=..\egg\configs

set LUBAN_DLL=%WORKSPACE%\Luban\Luban.dll
set CONF_ROOT=%WORKSPACE%

dotnet %LUBAN_DLL% ^
    -t server ^
    -c go-json ^
    -d json  ^
    --conf %CONF_ROOT%\luban.conf ^
    -x outputCodeDir=.\internal\gamedata ^
    -x outputDataDir=.\data\static ^
    -x pathValidator.rootDir=.\data\static ^
    -x l10n.provider=default ^
    -x l10n.textFile.path=Language@%WORKSPACE%\Datas\l10n\Language.xlsx ^
    -x l10n.textFile.keyFieldName=key ^
    -x l10n.textFile.languageFieldName=zh ^
    -x l10n.convertTextKeyToValue=1 ^
    -x lubanGoModule=demo/luban
pause