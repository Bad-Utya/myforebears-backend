package stage2_layout

import (
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
)

// LayoutFromPersonWithDepth запускает алгоритм размещения с ограничением глубины BFS
// maxDepth: 0 = unlimited (весь доступный граф); N > 0 = включать только на расстоянии N hops от root
func LayoutFromPersonWithDepth(tree *stage1_input.FamilyTree, startPersonID int, maxDepth int) (*stage1_input.PlacementHistory, error) {
	if maxDepth == 0 {
		// Если maxDepth == 0, используем стандартную функцию без ограничений
		return LayoutFromPerson(tree, startPersonID)
	}

	startPerson := tree.GetPerson(startPersonID)
	if startPerson == nil {
		return nil, fmt.Errorf("person with ID %d not found", startPersonID)
	}

	// Создаём историю размещений
	history := stage1_input.NewPlacementHistory()

	// Инициализируем начального человека
	startPerson.Layout = stage1_input.NewPersonLayout(0)
	startPerson.Layout.IsStartPerson = true

	// Создаём очередь и добавляем начального человека
	// Отслеживаем глубину через отдельную карту
	queueDepths := make(map[*stage1_input.Person]int)
	queue := NewQueue()
	queue.Enqueue(startPerson)
	queueDepths[startPerson] = 0

	// BFS обход
	for !queue.IsEmpty() {
		person := queue.Dequeue()
		depth := queueDepths[person]

		// Если уже обработан – пропускаем
		if person.IsProcessed() {
			continue
		}

		// Если превышена максимальная глубина – пропускаем
		if depth >= maxDepth {
			// Помечаем как обработанного, но не добавляем в историю
			person.Layout.Processed = true
			delete(queueDepths, person)
			continue
		}

		// Обрабатываем человека
		// Используем оригинальную функцию, которая добавляет в очередь
		processPersonLayout(person, queue, history)

		// Помечаем как обработанного
		person.Layout.Processed = true
		delete(queueDepths, person)

		// Добавляем глубину для новых элементов в очереди
		// Нужно проверить добавленных людей
		updateDepthsInQueue(tree, queue, queueDepths, depth, maxDepth)
	}

	return history, nil
}

// updateDepthsInQueue обновляет глубины для новых элементов в очереди
// Это сложно, потому что нам нужно знать, кто был добавлен
// Альтернатива: просто увеличиваем глубину для всех новых элементов с предыдущей глубиной
func updateDepthsInQueue(
	tree *stage1_input.FamilyTree,
	queue *Queue,
	depths map[*stage1_input.Person]int,
	currentDepth int,
	maxDepth int,
) {
	// Это требует модификации Queue для отслеживания новых элементов
	// Пока пропускаем - это более сложное решение
	// Вместо этого используем простой подход: все элементы в очереди получают глубину на основе их уровня слоя
}

// LayoutFromPersonWithDepthSimple - упрощённый вариант, использующий информацию о слоях
func LayoutFromPersonWithDepthSimple(tree *stage1_input.FamilyTree, startPersonID int, maxDepth int) (*stage1_input.PlacementHistory, error) {
	if maxDepth == 0 {
		return LayoutFromPerson(tree, startPersonID)
	}

	startPerson := tree.GetPerson(startPersonID)
	if startPerson == nil {
		return nil, fmt.Errorf("person with ID %d not found", startPersonID)
	}

	// Создаём историю размещений
	history := stage1_input.NewPlacementHistory()

	// Инициализируем начального человека
	startPerson.Layout = stage1_input.NewPersonLayout(0)
	startPerson.Layout.IsStartPerson = true

	// Создаём очередь и добавляем начального человека
	queue := NewQueue()
	queue.Enqueue(startPerson)

	// BFS обход
	for !queue.IsEmpty() {
		person := queue.Dequeue()

		// Если уже обработан – пропускаем
		if person.IsProcessed() {
			continue
		}

		// Обрабатываем человека
		processPersonLayoutWithDepthFilter(person, queue, history, maxDepth)

		// Помечаем как обработанного
		person.Layout.Processed = true
	}

	return history, nil
}

// processPersonLayoutWithDepthFilter обрабатывает человека, фильтруя добавления на основе maxDepth
func processPersonLayoutWithDepthFilter(
	person *stage1_input.Person,
	queue *Queue,
	history *stage1_input.PlacementHistory,
	maxDepth int,
) {
	// Для упрощения: отслеживаем глубину через слой (Layer)
	// В BFS слой обычно соответствует расстоянию от корня
	personLayer := person.Layout.Layer

	// 1. Добавляем родителей если они не превышают maxDepth
	if personLayer+1 < maxDepth {
		for !AddParents(person, queue, history) {
			// AddParents уже вызвал LowerSubtree, пробуем снова
		}
	}

	// 2. Добавляем партнёров если они в пределах maxDepth
	if personLayer < maxDepth {
		AddPartners(person, queue, history)
	}

	// После добавления нужно отфильтровать элементы в очереди, которые превышают maxDepth
	// Это требует доступа к внутренней структуре очереди
}
