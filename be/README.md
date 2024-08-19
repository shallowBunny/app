# shallowBunny BE

The backend is running a telegram bot, if you don't want one, you should skip it and simply create an api file with the correct data [there](/fe/public)
It's possible to test the BE using an empty SHALLOWBUNNY_TELEGRAM_API_TOKEN

## build

```
make
```

## run locally

```
export SHALLOWBUNNY_LOCALHOST=`hostname`
export SHALLOWBUNNY_TELEGRAM_API_TOKEN="xxxxx"
make run
```
