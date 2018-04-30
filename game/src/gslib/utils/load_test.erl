%% The MIT License (MIT)
%%
%% Copyright (c) 2014-2024
%% Savin Max <mafei.198@gmail.com>
%%
%% Permission is hereby granted, free of charge, to any person obtaining a copy
%% of this software and associated documentation files (the "Software"), to deal
%% in the Software without restriction, including without limitation the rights
%% to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
%% copies of the Software, and to permit persons to whom the Software is
%% furnished to do so, subject to the following conditions:
%%
%% The above copyright notice and this permission notice shall be included in all
%% copies or substantial portions of the Software.
%%
%% THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
%% IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
%% FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
%% AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
%% LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
%% OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
%% SOFTWARE.

-module(load_test).

-export([once/0, 
         on/1, 
         b/3, 
         bench/2, 
         bench/3, 
         summary/0, 
         d/1]).

-define(TAB, ?MODULE).
-define(CLOSE_SOCKET, true).

%% C: 并发客户端数量
%% N: 每个客户端发送请求数量
%% I: 客户端请求发送间隔时间

on(Clients) ->
    b(Clients, 10000, 1).

d(Clients) ->
    b(Clients, 10, 300).

once() ->
    b(1, 1, 1).

t(Requests, Delay) ->
    b(999, Requests, Delay).

parse_session() ->
    Path = "/Users/savin/code_fun/Golang/gos/sessions.txt",
    {ok, Content} = file:read_file(Path),
    Lines = string:tokens(binary_to_list(Content), "\n"),
    lists:foldl(fun(Line, Acc) ->
        [AccountId, Token] = string:tokens(Line, ","),
        [{list_to_binary(AccountId), list_to_binary(Token)}|Acc]
    end, [], Lines).

b(Clients, Requests, Delay) ->
    case ets:info(?TAB) of
        undefined -> do_nothing;
        _ -> ets:delete(?TAB)
    end,
    ets:new(?TAB, [set, public, named_table]),
    ets:insert(?TAB, {count, 0}),
    ets:insert(?TAB, {sent, 0}),
    ets:insert(?TAB, {msecs, 0}),
    ets:insert(?TAB, {c, Clients}),
    ets:insert(?TAB, {n, Requests}),
    ets:insert(?TAB, {error, 0}),
    ets:insert(?TAB, {number, 0}),
    Sessions = parse_session(),
    times(Clients, fun(Idx) -> 
        spawn(load_test, bench, [Requests, Delay, lists:nth(Idx, Sessions)]) 
    end).

times(0, _F) -> ok;
times(N, F) ->
    F(N),
    times(N - 1, F).

bench(N, I) ->
    bench(N, I, undefined).

bench(N, I, Session) ->
    Sock = connect(62369),
    Udid = "tt",
    StartTimeStamp = os:timestamp(),

    if
        Session =:= undefined ->
            %% Login
            Protocol = <<1:16>>,
            AccountId  = <<"usernamet10">>,
            AccountIdLen = byte_size(AccountId),
            Token  = <<"tISWFanBJkKHuxjOmlNSuSkwOQAzlzOV">>,
            TokenLen = byte_size(Token),
            gen_tcp:send(Sock, list_to_binary([Protocol, 
                                               <<AccountIdLen:16>>, AccountId,
                                               <<TokenLen:16>>, Token]));
        true ->
            %% Login
            Protocol = <<1:16>>,
            {AccountId, Token}  = Session,
            AccountIdLen = byte_size(AccountId),
            TokenLen = byte_size(Token),
            gen_tcp:send(Sock, list_to_binary([Protocol, 
                                               <<AccountIdLen:16>>, AccountId,
                                               <<TokenLen:16>>, Token]))
    end,

    {ok, Res} = gen_tcp:recv(Sock, 0),
    error_logger:info_msg("LoginRes: ~p~n", [Res]),

    run(N, I, Sock, Udid),
    StopTimeStamp = os:timestamp(),
    result(StartTimeStamp, StopTimeStamp),
    exit(normal).

connect(Port) ->
    SomeHostInNet = "127.0.0.1", % to make it runnable on one machine
    case gen_tcp:connect(SomeHostInNet, Port, [{active, false}, {packet, 2}]) of 
        {ok, Socket} -> Socket;
        {error, Reason} -> 
            error_logger:info_msg("connect error: ~p~n", [Reason])
    end.

run(0, _I, Sock, _Udid) -> 
    case ?CLOSE_SOCKET of
        true ->
            gen_tcp:close(Sock);
        false -> do_nothing
    end;
run(N, I, Sock, Udid) ->
    if
        I > 0 ->
            timer:sleep(I);
        true ->
            do_nothing
    end,

    gen_tcp:send(Sock, list_to_binary([<<5:16>>, 
                                       <<1:16>>, <<"a">>,
                                       <<1:16>>, <<"b">>,
                                       <<1:16>>, <<"c">>])),

    ets:update_counter(?TAB, sent, 1),
    case gen_tcp:recv(Sock, 0) of
        {ok, Packet} -> 
            % error_logger:info_msg("Response: ~p~n", [Packet]),
            ets:update_counter(?TAB, count, 1),
            run(N-1, I, Sock, Udid);
        Error -> 
            case ?CLOSE_SOCKET of
                true ->
                    gen_tcp:close(Sock);
                false -> do_nothing
            end,
            error_logger:info_msg("error: ~p~n", [Error]),
            ets:update_counter(?TAB, error, 1)
    end.

result(StartTimeStamp, StopTimeStamp) ->
    {_StartMegaSecs, StartSecs, StartMicroSecs} = StartTimeStamp,
    {_StopMegaSecs, StopSecs, StopMicroSecs} = StopTimeStamp,
    UsedMicroSecs = StopMicroSecs - StartMicroSecs + (StopSecs - StartSecs) * 1000000,
    ets:update_counter(?TAB, msecs, UsedMicroSecs),
    ets:update_counter(?TAB, number, 1),
    io:format("ok~n").

summary() ->
    [{c, C}] = ets:lookup(?TAB, c),
    [{n, N}] = ets:lookup(?TAB, n),
    [{error, Error}] = ets:lookup(?TAB, error),
    [{count, Count}] = ets:lookup(?TAB, count),
    [{sent, Sent}]  = ets:lookup(?TAB, sent),
    [{msecs, UsedMicroSecs}] = ets:lookup(?TAB, msecs),
    [{number, FinishedClientCount}] = ets:lookup(?TAB, number),
    io:format("Used: ~pus~n", [UsedMicroSecs / FinishedClientCount]),
    io:format("Used: ~ps~n", [UsedMicroSecs / 1000000 / FinishedClientCount]),
    MicroSecsPerRequest = UsedMicroSecs / Count / FinishedClientCount,
    io:format("FinishedClientCount: ~p~n", [FinishedClientCount]),
    io:format("Total Request: ~p~n", [C * N]),
    io:format("Successed Request: ~p~n", [Count]),
    io:format("Sent Request: ~p~n", [Sent]),
    io:format("Error Request: ~p~n", [Error]),
    io:format("MicroSecsPerRequest: ~p~n", [MicroSecsPerRequest]),
    io:format("Requests Per Seconds: ~p~n", [1000000 / MicroSecsPerRequest]).
