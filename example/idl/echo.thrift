namespace go echo

struct Request {
    1: string message;
}

service Echo {
    oneway void Ping(1: Request request);
}

