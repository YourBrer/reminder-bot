import { MobileDateTimePicker } from '@mui/x-date-pickers/MobileDateTimePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { useState } from "react";
import styled from "@emotion/styled";
import type { PickerValue } from "@mui/x-date-pickers/internals";
import 'dayjs/locale/ru';
import { renderTimeViewClock } from "@mui/x-date-pickers";

const Container = styled.div({
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
});
const ButtonContainer = styled.div({
    width: '80%',
    display: 'flex',
    justifyContent: 'space-between',
    gap: '10px',
    marginTop: '10px',
});
const Button = styled.button({
    border: '1px solid #007aff',
    color: '#007aff',
    borderRadius: '8px',
    padding: '8px',
    fontWeight: 'bold',
    backgroundColor: '#e8f2ff',
    transition: 'background-color ease 0.2s',
    '&:hover': {
        backgroundColor: '#cbd1d9',
        cursor: 'pointer',
    },

    '&.submit': {
        color: '#fff',
        backgroundColor: '#007aff',
        '&:hover': {
            backgroundColor: '#0357b2',
        },
    }
});
const Error = styled.p({
    color: 'red',
    marginBottom: '10px',
});

const webAppClose = window.Telegram.WebApp.close;

function App() {
    const [date, setDate] = useState<PickerValue | null>(null);
    const [hasError, toggleError] = useState(false);

    function onSubmit() {
        if (hasError) return;
        if (!date) {
            toggleError(true);
            return;
        }

        const taskId = location.search
            ? location.search.replace('?taskId=', '')
            : null;
        if (taskId) {
            fetch(
                import.meta.env.VITE_URL,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ taskId, date: date.format() }),
                },
            )
                .then(() => alert('Дата установлена'))
                .catch(() => alert('Не удалось установить дату'))
                .finally(webAppClose)
        } else {
            webAppClose();
        }
    }

    return <Container>
        {hasError && <Error>Выберите дату и время</Error>}

        <LocalizationProvider dateAdapter={AdapterDayjs} adapterLocale='ru'>
            <MobileDateTimePicker
                value={date}
                ampm={false}
                disablePast
                format="DD.MM.YY HH:mm"
                label='Выбери дату и время'
                timeSteps={{
                    hours: 1,
                    minutes: 10,
                }}
                onChange={setDate}
                onAccept={value => toggleError(!value)}
                viewRenderers={{
                    hours: renderTimeViewClock,
                    minutes: renderTimeViewClock,
                }}
            />
        </LocalizationProvider>
        <ButtonContainer>
            <Button onClick={webAppClose}>Отмена</Button>
            <Button className='submit' onClick={onSubmit}>Готово</Button>
        </ButtonContainer>
    </Container>
}

export default App
