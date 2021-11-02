#include <stdlib.h>
#include <stdio.h>
#include <string.h>

char *hello(char *name)
{
    char *out = malloc(sizeof(char) * 5 + strlen(name));
    sprintf(out, "Hi %s!", name);
    return out;
}

int main()
{
    char *greeting = hallo("friday");
    printf("%s\n", greeting);
    free(greeting);
    return 0;
}