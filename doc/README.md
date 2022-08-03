# Hosting Framework Overview



Hosting framework helps developers managing their application resources and easy their life with host, service and component concepts. It provides generic components to encapsulate application resources and  functionalities, and it also provides unified approach to manage components involved in application with interfaces and configure/factory methods. As components are managed with dependency injection, developers can easily access existing components or define their own components to customize the application's behavior.



## Design Reference

The design has referenced dot net runtime libraries including hosting & dependency injection. Check it out to understand the high level concepts.



Below is the link to docs of relevant runtime libraries in dot net:

[.NET Generic Host | Microsoft Docs](https://docs.microsoft.com/en-us/dotnet/core/extensions/generic-host)

[Dependency injection in .NET | Microsoft Docs](https://docs.microsoft.com/en-us/dotnet/core/extensions/dependency-injection)



## Usage

### Two Sets of APIs

The framework provides two sets of APIs for developers to use. you can choose either of them for your purpose:

- Integrated API: Hosting

  Hosting API is similar to the Microsoft.Extensions.Hosting framework, which encapsulate your whole application as a Host and runs services inside the host. It is integrated with the whole design of your application. You will need to understand the concepts like Host, Service, Components, configuration, logging, etc. you will also leverage the framework API to run the application via concept like Service, AppRunner, Looper and Processor, etc.

  For new applications, you should use this API to fully adopt the Dependency Injection pattern and leverage the advantage of it.

- Standalone API: Activator

  Standalone API does not involve deeply on how your application will be executed, but focus on how components are registered and inject dependency as needed. it manages only the dependency of components in your application. In this API we don't provide concept like host, service, looper, processor. Instead, dependency types, lifecycle, scope are the major focus we can help for your application. 

  You can use this API for legacy code, thus no need to refactor the whole of your code.

Both these two sets of API share largely the same implementation underneath on the dependency management and injection functionalities. the concept is also the same on DI and IoC.

Sections:

- [Shared APIs for Dependency Registration and Injection](./API/CommonAPI.md)
- [Standalone APIs: Activator](./API/StandaloneAPI.md)
- [Integrated APIs: Hosting](./API/HostingAPI.md)

### Example

Here is an hello world example to show how to use the framework:

[Helloworld](./samples/Helloworld.md)

[Console Application](./samples/Console%20Application.md)

[Activator Sample](./samples/Activator%20Sample.md)

### Diagnostic Support

Tools:

- [Diagnose Issues](./howto/DiagnoseIssues.md)
- [Cyclic Dependency Detection](./concepts/DependencyInjection.md)



## Topics

- [[Architecture](./Architecture.md)]
- [Concepts](./concepts/README.md)
  1. [Component](./Component.md)
  2. [Service](./Service.md)
  3. [Host](./concepts/Host.md)
  4. [Factory Method](./concepts/FactoryMethod.md)
  7. [AppRunner](./concepts/AppRunner.md)
  8. [Configuration](./concepts/Configuration.md)
  9. [Logging](./concepts/Logging.md)
  10. [Looper](./Looper.md)
- Samples
  - [HelloWorld](./Helloworld.md)
  - [Console Application](./samples/Console Application.md)
- How To
  - Register a Component
  - Get and Use a Dependent Component
  - Create a Service
  - [Create Window Service](./howto/WindowsService.md)
  - [Create a Looper](./concepts/Looper.md)
  - Create a Loop Processor
  - Register for Lifecycle callbacks
- Advanced
  - [Understand Dependency Injection](./concepts/DependencyInjection.md)
  - [Understand the Context](./concepts/Context.md)
- Q&A