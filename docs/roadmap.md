## Roadmap

This is an overview of the roadmap.
The project is managed on Github's [lib-bpmn-engine milestones](https://github.com/nitram509/lib-bpmn-engine/milestones) page.

#### ✅ v0.1.0

[progress milestone v0.1.0](///github.com/nitram509/lib-bpmn-engine/issues?q=is%3Aissue+milestone%3Av0.1.0+is%3Aclosed)

For the first release I would like to have service tasks and events fully supported.


#### ✅ v0.2.0

[progress milestone v0.2.0](///github.com/nitram509/lib-bpmn-engine/issues?q=is%3Aissue+milestone%3Av0.2.0+is%3Aclosed)

With basic element support, I would like to add [visualization/monitoring](./advanced-zeebe.md) capabilities.
If the idea of using Zeebe's exporter protocol is not too complex, that would be ideal.
If not, a simple console logger might do the job as well.
Also, I would like to add [expression language support](./expression-syntax.md) as well as support for correlation keys


#### ⚙️ v0.3.0

[progress milestone v0.3.0](///github.com/nitram509/lib-bpmn-engine/issues?q=is%3Aissue+milestone%3Av0.3.0)

One last but very important feature I aim for is the ability to load & store state.
Which means, that you as app developer would be able to persistent in-flight process instances
for later restoring and completion.

#### 🔮️ v0.?.0

🤔more elements to be supported ... or more events to be exported
