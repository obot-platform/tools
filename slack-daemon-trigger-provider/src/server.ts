import express, { Request, Response, NextFunction } from 'express'
import bodyParser from 'body-parser'

export const startServer = (port: number) => {
    const app = express()
    app.use(bodyParser.json())

    // TODO(njhale): Remove this after local testing
    app.use((req: Request, res: Response, next: NextFunction) => {
        console.log('--- Incoming HTTP Request ---')
        console.log(`Method: ${req.method}`)
        console.log(`URL: ${req.url}`)
        console.log('Headers:', req.headers)
        console.log('Body:', JSON.stringify(req.body, null, 2))
        console.log('--------------------------------')
        next()
    })

    app.get('/', (_req: Request, res: Response) => res.send('OK'))

    app.post('/*', (req: Request, res: Response) => {
        console.log('POST Request:', req.body)
        res.send('OK')
    })

    return app.listen(port, '127.0.0.1', () => {
        console.log(`Server is listening on port ${port}`)
    })
}
