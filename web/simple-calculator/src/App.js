import React, { useState } from "react";
import "./App.css";

function App() {
  const [expression, setExpression] = useState(""); // Хранение текущего ввода
  const [result, setResult] = useState(null); // Хранение результата
  const [error, setError] = useState(null); // Хранение ошибок

  const handleButtonClick = (value) => {
    setExpression((prevExpression) => prevExpression + value);
  };

  const handleClear = () => {
    setExpression("");
    setResult(null);
    setError(null);
  };

  const handleCalculate = async () => {
    try {
      const response = await fetch("http://127.0.0.1:8080/api/v1/calculate", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ expression }), // Отправляем выражение
      });
  
      const data = await response.json();
      if (response.ok && data.id) {
        fetchResult(data.id); // Если запрос успешен, запрашиваем результат по ID
      } else {
        setError("Ошибка вычисления");
      }
    } catch (error) {
      console.error("Error calculating expression:", error);
      setError("Ошибка при отправке запроса");
    }
  };
  
  const fetchResult = async (id) => {
    try {
      // Запрашиваем результат по ID
      const response = await fetch(`http://127.0.0.1:8080/api/v1/expressions/${id}`);
      const data = await response.json();

      if (data.expression && data.expression.result !== null) {
        // Если результат найден, обновляем состояние
        setResult(data.expression.result);
      } else {
        // Если результат ещё не готов, проверяем через секунду
        setTimeout(() => fetchResult(id), 1000);
      }
    } catch (error) {
      console.error("Error fetching result:", error);
      setError("Ошибка при получении результата");
    }
  };

  return (
    <div className="App">
      <h1>Simple Calculator</h1>
      <div className="calculator">
        <input
          type="text"
          value={expression}
          readOnly
          className="calculator-screen"
        />
        <div className="button-container">
          <button onClick={() => handleButtonClick("1")}>1</button>
          <button onClick={() => handleButtonClick("2")}>2</button>
          <button onClick={() => handleButtonClick("3")}>3</button>
          <button onClick={() => handleButtonClick("+")}>+</button>
          <button onClick={() => handleButtonClick("4")}>4</button>
          <button onClick={() => handleButtonClick("5")}>5</button>
          <button onClick={() => handleButtonClick("6")}>6</button>
          <button onClick={() => handleButtonClick("-")}>-</button>
          <button onClick={() => handleButtonClick("7")}>7</button>
          <button onClick={() => handleButtonClick("8")}>8</button>
          <button onClick={() => handleButtonClick("9")}>9</button>
          <button onClick={() => handleButtonClick("*")}>*</button>
          <button onClick={() => handleButtonClick("0")}>0</button>
          <button onClick={handleClear}>C</button>
          <button onClick={handleCalculate}>=</button>
          <button onClick={() => handleButtonClick("/")}>/</button>
        </div>
        {error && (
          <div className="error">
            <h2>Error: {error}</h2>
          </div>
        )}
        {result !== null && (
          <div className="result">
            <h2>Result: {result}</h2>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
