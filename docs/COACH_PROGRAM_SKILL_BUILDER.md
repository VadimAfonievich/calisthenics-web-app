# Coach Program and Skill Builder

## Content hierarchy

The authoring flow uses the existing model:

1. Exercises are reusable movements.
2. Workouts contain ordered exercises with sets, targets, and rest.
3. Programs contain ordered program stages; every stage owns its ordered workouts.
4. Skills appear in a Student Skill Map category and contain their own progression stages.
5. A skill stage may link to one program stage. The linked workouts then appear inside that skill stage.
6. Skill prerequisites are separate from both the Skill Map category and the internal progression stages.

## Recommended authoring order

Create and publish exercises, then workouts, then a multi-stage program, and finally the skill. Publication checks prevent a program from using unpublished workouts and prevent a skill from using an unpublished prerequisite or program.

## Example: Path to Planche

Create program `Путь к горизонту` with stages such as `Подготовка прямых рук`, `Tuck Planche`, `Straddle Planche`, and `Full Planche`. Assign the relevant workouts to each stage.

Create one skill `Горизонт` in category `Пол`. Add the matching progression stages and select the corresponding `Путь к горизонту · <этап>` in each skill stage. Do not create Tuck, Straddle, and Full Planche as independent top-level Skill Map nodes unless they are intentionally separate product skills.

Existing programs created before multi-stage authoring remain compatible and open as a single program stage.
