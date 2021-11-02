# error-gallery

Show how different compilers tell you that you called a function `hallo`
that doesn't exist.

[Live Demo](https://flofriday.github.io/error-gallery/)

## Generate the index file

```bash
docker image build --tag error-gallery .
docker run -it --rm --volume $PWD:/PWD error-gallery
```
