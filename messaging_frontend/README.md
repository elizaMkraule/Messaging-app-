## Getting stated

Install all of the required dependencies in order to build your project:

```npm install```

This will automatically install all of the tools and dependencies that are
needed and allowed for the project.  You should only need to run npm install
once each time you create a fresh clone.

You should *not* use `npm install` to install anything else. The packages in
`package.json` that are installed with this command are the only packages you
are allowed to use.
## Build

Before you build, you will need to create an empty "schemas" directory. This
will be the directory that you will eventually add your schemas to.

After you have created that directory, you can build your project as follows:

```npm run build```

This will create a file called "dist/index.html" along with the necessary
additional files.  If you open "dist/index.html" in a web browser you will see
your application.

As you develop, you may instead want to use:

```npm start```

This starts up a web server and tells you the correct URL.  If you navigate to
that URL, you will see your application.  The advantage of this method is that
it does "hot module reloading".  This means that as you edit and save your
files, they will automatically be reloaded in the web browser.  Note, however,
that this will not always work seemlessly, as you will lose some application
state, depending on what you edited. So, you still may need to hit "refresh" in
your browser.
