@setlocal

go test -c ./test
set GODEBUG=allocfreetrace=1
test.test -test.run=none -test.bench=BenchmarkRog -test.benchtime=10ms 2>trace.log

@endlocal
exit /b 0
