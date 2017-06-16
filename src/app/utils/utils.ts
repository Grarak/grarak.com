export class Utils {
    static getAsset(file: string): string {
        return 'assets/' + file;
    }

    static formatSeconds(secs: number): string {
        const hours = Math.floor(secs / 60 / 60)
        const minutes = Math.floor(secs / 60) % 60
        const seconds = Math.ceil(secs % 60)

        let hText = hours.toString()
        let mText = minutes.toString()
        let sText = seconds.toString()
        if (hours < 10) {
            hText = '0' + hText
        }
        if (minutes < 10) {
            mText = '0' + mText
        }
        if (seconds < 10) {
            sText = '0' + sText
        }

        return hText + ':' + mText + ':' + sText
    }
}
