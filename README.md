<br>

<div align="center" markdown=1>
  <p align="center"><img src="https://i.imgur.com/CzXM1jx.png" style="width: 200px"></p>

  <a href='https://github.com/redi-db/redi.db/tree/main#-getting-started'>Getting Started</a> • 
  <a href="https://github.com/redi-db/redi.db/tree/main#%EF%B8%8F-installing">Installing</a> •
  <a href='https://github.com/redi-db/redi.db/tree/main#%EF%B8%8F-how-do-i-compile-the-project-myself'>How compile</a>
</div>

<br>

<h1 align="center">🎉 Getting Started</h1>
Redi.db is a large and powerful data repository that is developed in the GoLang programming language. It provides an efficient, fast and reliable way to store, process and manage data.

One of the main advantages of redi.db is its speed. This database has high performance and speed, which allows you to process huge amounts of data in a short time. This makes redi.db suitable for projects with high load, where speed plays an important role.

In addition, redi.db provides many features that make it even more useful and easy to use. For example, it supports several data types, including strings, lists, hashes, and sets. This allows it to work flexibly with different data types, making it an ideal choice for a variety of projects.

It is also worth noting that redi.db has a high degree of reliability and robustness. It uses various mechanisms to ensure data integrity, including data replication and disaster recovery mechanisms. This allows it to guarantee data security and integrity under all conditions.

Finally, redi.db provides a user-friendly interface for working with data. It has a simple and intuitive API that integrates easily into various projects. This makes it an ideal choice for developers who want to quickly and easily integrate the database into their projects.

Overall, redi.db is a powerful and efficient tool for storing, processing, and managing data. Its speed, flexibility, reliability, and ease-of-use make it an ideal choice for a wide range of projects, from small applications to large systems.

<hr>

The Redi.db database is a free, fast, and lightweight database in the Go programming language that offers many features for easy data manipulation. One of the most interesting features is the ability to create custom data lookup queries.

How it works: You can define your own data search query using JSON-based query language. Search queries can be performed at several levels, including keyword searches, search queries using the logical operators and pattern searches.

<hr>

In addition, the Redi.db database supports connection through the HTTP and WebSockets protocols. This means that you can access the database both through a browser and through other client applications, using special libraries and frameworks.

Using the HTTP protocol allows you to access the database through a RESTful API. You can do this by sending requests to a specific URL, specifying an HTTP method (e.g. GET, POST, PUT, DELETE) and request data according to API requirements.

Using WebSockets allows you to establish a persistent connection between the client and the server, allowing for faster and more efficient data exchange. Clients can send requests and receive responses in real time without reloading the page or establishing a new connection.

Overall, the Redi.db database offers many features that make it very attractive for various projects. It is lightweight, fast, supports custom search queries, and has support for HTTP and WebSockets protocols. If you are looking for a simple and efficient way to store data, the Redi.db database may be a great choice for you

<br>

<h1 align="center">⬇️ Installing</h1>
To download the finished, compiled version of the database, you can go to <a href="https://github.com/redi-db/redi.db/releases">releases</a> and download the latest version for your OS and architecture.

<br>

<h3>Installation steps:</h3>
- Download the archive for your operating system<br>
- Select and retrieve the desired architecture<br>
- Open the executable file.<br>

<br>

After running the file, config.yml will appear in the current path with the settings of the available parameters. (each item is signed, it is impossible to make a mistake)

<h4>🔺 If you are running on a Linux operating system and we get an error - you need to give rights.</h4>
<pre>chmod -R 777 ./RediDB-x</pre>

<br>

<h1 align="center">⚙️ How do I compile the project myself?</h1>
In order to build a project on our own, we need to download the source files. You can do this as in the <a href="https://github.com/redi-db/redi.db/releases">release</a> or by <a href="https://github.com/redi-db/redi.db/archive/refs/heads/main.zip">clicking on me =)</a>.

To compile, we need golang, which you can <a href="https://go.dev/dl/">download here</a>. <br>

Next, extract all the files into any folder. Open a terminal in this folder and write the following. (one by one)

<pre>go mod tidy</pre>

And at the end use these commands: <b>for linux:</b> `./build.sh`, <b>for windows:</b> `./build.bat`.

After compilation, a bin folder will appear in the current path with executables (for different versions and architectures).
<br>
