# Hosting Framework Overview



Hosting framework helps developers managing their application resources and easy their life with host, service and component concepts. It provides generic components to encapsulate application resources and  functionalities, and it also provides unified approach to manage components involved in application with interfaces and configure/factory methods. As components are managed with dependency injection, developers can easily access existing components or define their own components to customize the application's behavior.



## Design Reference

The design has referenced dot net runtime libraries including hosting & dependency injection. Check it out to understand the high level concepts.



Below is the link to docs of relevant runtime libraries in dot net:

[.NET Generic Host | Microsoft Docs](https://docs.microsoft.com/en-us/dotnet/core/extensions/generic-host)

[Dependency injection in .NET | Microsoft Docs](https://docs.microsoft.com/en-us/dotnet/core/extensions/dependency-injection)



## Usage

Here is an hello world example to show how to use the framework:

[Helloworld](./samples/Helloworld.md)



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