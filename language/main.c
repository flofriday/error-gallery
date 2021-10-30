#include <stdio.h>
#include <stdlib.h>

char *hello(char *name)
{
    char *out;
    asprintf(&out, "Hi %s!", name);
    return out;
}

int main()
{
    char *greeting = hallo("friday");
    printf("%s\n", greeting);
    free(greeting);
    return 0;
}